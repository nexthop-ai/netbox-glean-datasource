// Copyright (c) 2026-present, Nexthop Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package netbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	pageSize   int
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		pageSize: 1000,
	}
}

// SetHTTPClient replaces the default HTTP client (useful for testing).
func (c *Client) SetHTTPClient(hc *http.Client) {
	c.httpClient = hc
}

type ListResponse struct {
	Count    int              `json:"count"`
	Next     *string          `json:"next"`
	Previous *string          `json:"previous"`
	Results  []map[string]any `json:"results"`
}

// List fetches all pages from a NetBox API endpoint, calling fn for each page of results.
func (c *Client) List(ctx context.Context, endpoint string, params url.Values, fn func(results []map[string]any) error) error {
	if params == nil {
		params = url.Values{}
	}
	if params.Get("limit") == "" {
		params.Set("limit", strconv.Itoa(c.pageSize))
	}

	reqURL := c.baseURL + endpoint + "?" + params.Encode()

	for reqURL != "" {
		resp, err := c.doWithRetry(ctx, reqURL)
		if err != nil {
			return err
		}

		var page ListResponse
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			_ = resp.Body.Close()
			return fmt.Errorf("decoding response from %s: %w", endpoint, err)
		}
		_ = resp.Body.Close()

		if len(page.Results) > 0 {
			if err := fn(page.Results); err != nil {
				return err
			}
		}

		if page.Next != nil && *page.Next != "" {
			reqURL = *page.Next
			// If the server returns an absolute URL with a different host (e.g., internal hostname),
			// rewrite it to use our configured base URL.
			if parsed, err := url.Parse(reqURL); err == nil {
				base, _ := url.Parse(c.baseURL)
				if parsed.Host != base.Host {
					parsed.Scheme = base.Scheme
					parsed.Host = base.Host
					reqURL = parsed.String()
				}
			}
		} else {
			reqURL = ""
		}
	}
	return nil
}

const maxRetries = 3

func (c *Client) doWithRetry(ctx context.Context, reqURL string) (*http.Response, error) {
	for attempt := range maxRetries {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Token "+c.token)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				slog.Warn("request failed, retrying", "url", reqURL, "attempt", attempt+1, "error", err)
				sleep(ctx, backoff(attempt))
				continue
			}
			return nil, fmt.Errorf("request to %s: %w", reqURL, err)
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			return resp, nil
		case resp.StatusCode == http.StatusTooManyRequests:
			_ = resp.Body.Close()
			wait := parseRetryAfter(resp.Header.Get("Retry-After"), backoff(attempt))
			slog.Warn("rate limited, waiting", "url", reqURL, "wait", wait)
			sleep(ctx, wait)
			continue
		case resp.StatusCode >= 500:
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if attempt < maxRetries-1 {
				slog.Warn("server error, retrying", "url", reqURL, "status", resp.StatusCode, "attempt", attempt+1)
				sleep(ctx, backoff(attempt))
				continue
			}
			return nil, fmt.Errorf("request to %s returned %d: %s", reqURL, resp.StatusCode, string(body))
		default:
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return nil, fmt.Errorf("request to %s returned %d: %s", reqURL, resp.StatusCode, string(body))
		}
	}
	return nil, fmt.Errorf("request to %s failed after %d retries", reqURL, maxRetries)
}

func backoff(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}

func parseRetryAfter(header string, fallback time.Duration) time.Duration {
	if header == "" {
		return fallback
	}
	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if t, err := time.Parse(time.RFC1123, header); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return fallback
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// Helper functions for extracting typed values from map[string]any.

func GetString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func GetInt(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return 0
	}
}

func GetFloat64(m map[string]any, key string) float64 {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}

func GetNested(m map[string]any, keys ...string) map[string]any {
	current := m
	for _, key := range keys {
		if current == nil {
			return nil
		}
		v, ok := current[key]
		if !ok || v == nil {
			return nil
		}
		nested, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		current = nested
	}
	return current
}

func GetNestedString(m map[string]any, keys ...string) string {
	if len(keys) == 0 {
		return ""
	}
	nested := GetNested(m, keys[:len(keys)-1]...)
	return GetString(nested, keys[len(keys)-1])
}

func GetTime(m map[string]any, key string) *time.Time {
	s := GetString(m, key)
	if s == "" {
		return nil
	}
	// NetBox uses ISO 8601 timestamps.
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05.999999Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

func GetTags(m map[string]any) []string {
	v, ok := m["tags"]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var tags []string
	for _, item := range arr {
		switch t := item.(type) {
		case map[string]any:
			if name := GetString(t, "display"); name != "" {
				tags = append(tags, name)
			} else if name := GetString(t, "name"); name != "" {
				tags = append(tags, name)
			}
		case string:
			tags = append(tags, t)
		}
	}
	return tags
}

func GetBool(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

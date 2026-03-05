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
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func TestListPagination(t *testing.T) {
	page := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		page++
		var resp ListResponse
		switch page {
		case 1:
			next := fmt.Sprintf("http://%s/api/dcim/devices/?limit=2&offset=2", r.Host)
			resp = ListResponse{
				Count: 3,
				Next:  &next,
				Results: []map[string]any{
					{"id": float64(1), "name": "device-1"},
					{"id": float64(2), "name": "device-2"},
				},
			}
		case 2:
			resp = ListResponse{
				Count: 3,
				Next:  nil,
				Results: []map[string]any{
					{"id": float64(3), "name": "device-3"},
				},
			}
		default:
			t.Fatalf("unexpected page request: %d", page)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")
	client.pageSize = 2

	var allResults []map[string]any
	err := client.List(context.Background(), "/api/dcim/devices/", nil, func(results []map[string]any) error {
		allResults = append(allResults, results...)
		return nil
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(allResults) != 3 {
		t.Fatalf("expected 3 results, got %d", len(allResults))
	}
	if GetString(allResults[2], "name") != "device-3" {
		t.Errorf("expected device-3, got %s", GetString(allResults[2], "name"))
	}
}

func TestListWithParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("last_updated__gte") != "2024-01-01T00:00:00Z" {
			t.Errorf("expected last_updated__gte param, got: %s", r.URL.RawQuery)
		}
		resp := ListResponse{Count: 0, Results: []map[string]any{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")
	params := url.Values{}
	params.Set("last_updated__gte", "2024-01-01T00:00:00Z")

	err := client.List(context.Background(), "/api/dcim/devices/", params, func(results []map[string]any) error {
		return nil
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestRetryOn5xx(t *testing.T) {
	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
			return
		}
		resp := ListResponse{Count: 1, Results: []map[string]any{{"id": float64(1)}}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")

	var results []map[string]any
	err := client.List(context.Background(), "/api/test/", nil, func(r []map[string]any) error {
		results = append(results, r...)
		return nil
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if int(attempts.Load()) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestRetryOn429(t *testing.T) {
	var attempts atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		resp := ListResponse{Count: 1, Results: []map[string]any{{"id": float64(1)}}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-token")

	var results []map[string]any
	err := client.List(context.Background(), "/api/test/", nil, func(r []map[string]any) error {
		results = append(results, r...)
		return nil
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestGetString(t *testing.T) {
	m := map[string]any{"name": "test", "count": float64(42), "nil_val": nil}

	if v := GetString(m, "name"); v != "test" {
		t.Errorf("expected 'test', got %q", v)
	}
	if v := GetString(m, "count"); v != "42" {
		t.Errorf("expected '42', got %q", v)
	}
	if v := GetString(m, "missing"); v != "" {
		t.Errorf("expected empty, got %q", v)
	}
	if v := GetString(m, "nil_val"); v != "" {
		t.Errorf("expected empty for nil, got %q", v)
	}
	if v := GetString(nil, "key"); v != "" {
		t.Errorf("expected empty for nil map, got %q", v)
	}
}

func TestGetInt(t *testing.T) {
	m := map[string]any{"id": float64(42), "name": "test"}

	if v := GetInt(m, "id"); v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
	if v := GetInt(m, "name"); v != 0 {
		t.Errorf("expected 0, got %d", v)
	}
	if v := GetInt(m, "missing"); v != 0 {
		t.Errorf("expected 0, got %d", v)
	}
}

func TestGetNestedString(t *testing.T) {
	m := map[string]any{
		"site": map[string]any{
			"display": "Site A",
			"region": map[string]any{
				"name": "US-East",
			},
		},
	}

	if v := GetNestedString(m, "site", "display"); v != "Site A" {
		t.Errorf("expected 'Site A', got %q", v)
	}
	if v := GetNestedString(m, "site", "region", "name"); v != "US-East" {
		t.Errorf("expected 'US-East', got %q", v)
	}
	if v := GetNestedString(m, "site", "missing"); v != "" {
		t.Errorf("expected empty, got %q", v)
	}
	if v := GetNestedString(m, "missing", "display"); v != "" {
		t.Errorf("expected empty for missing parent, got %q", v)
	}
}

func TestGetTime(t *testing.T) {
	m := map[string]any{
		"created":      "2024-06-15T10:30:00Z",
		"last_updated": "2024-06-15T10:30:00.123456+00:00",
		"date_only":    "2024-06-15",
		"invalid":      "not-a-date",
		"empty":        "",
	}

	if v := GetTime(m, "created"); v == nil {
		t.Error("expected non-nil time for 'created'")
	} else if v.Year() != 2024 || v.Month() != 6 || v.Day() != 15 {
		t.Errorf("unexpected time: %v", v)
	}

	if v := GetTime(m, "last_updated"); v == nil {
		t.Error("expected non-nil time for 'last_updated'")
	}

	if v := GetTime(m, "date_only"); v == nil {
		t.Error("expected non-nil time for 'date_only'")
	}

	if v := GetTime(m, "invalid"); v != nil {
		t.Errorf("expected nil for invalid date, got %v", v)
	}

	if v := GetTime(m, "empty"); v != nil {
		t.Errorf("expected nil for empty, got %v", v)
	}

	if v := GetTime(m, "missing"); v != nil {
		t.Errorf("expected nil for missing, got %v", v)
	}
}

func TestGetTags(t *testing.T) {
	// NetBox-style tags (array of objects).
	m := map[string]any{
		"tags": []any{
			map[string]any{"display": "production", "name": "production", "slug": "production"},
			map[string]any{"display": "core", "name": "core", "slug": "core"},
		},
	}
	tags := GetTags(m)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "production" || tags[1] != "core" {
		t.Errorf("unexpected tags: %v", tags)
	}

	// No tags.
	m2 := map[string]any{}
	if tags := GetTags(m2); tags != nil {
		t.Errorf("expected nil tags, got %v", tags)
	}

	// Null tags.
	m3 := map[string]any{"tags": nil}
	if tags := GetTags(m3); tags != nil {
		t.Errorf("expected nil tags, got %v", tags)
	}
}

func TestGetBool(t *testing.T) {
	m := map[string]any{"enabled": true, "disabled": false, "name": "test"}

	if v := GetBool(m, "enabled"); !v {
		t.Error("expected true")
	}
	if v := GetBool(m, "disabled"); v {
		t.Error("expected false")
	}
	if v := GetBool(m, "name"); v {
		t.Error("expected false for non-bool")
	}
	if v := GetBool(m, "missing"); v {
		t.Error("expected false for missing")
	}
}

func TestParseRetryAfter(t *testing.T) {
	fallback := 5 * time.Second

	if d := parseRetryAfter("", fallback); d != fallback {
		t.Errorf("expected fallback, got %v", d)
	}
	if d := parseRetryAfter("3", fallback); d != 3*time.Second {
		t.Errorf("expected 3s, got %v", d)
	}
	if d := parseRetryAfter("invalid", fallback); d != fallback {
		t.Errorf("expected fallback for invalid, got %v", d)
	}
}

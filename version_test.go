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

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	w := httptest.NewRecorder()

	handleVersion(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}

	var v versionInfo
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// The version info should have at least the revision field
	if v.Revision == "" {
		t.Error("expected non-empty revision")
	}

	// BuildTime should be set (even if "unknown")
	if v.BuildTime == "" {
		t.Error("expected non-empty buildTime")
	}
}

func TestReadBuildInfo(t *testing.T) {
	v := readBuildInfo()

	// Should always return a valid struct, even if values are "unknown"
	if v.Revision == "" {
		t.Error("expected non-empty revision")
	}
	if v.BuildTime == "" {
		t.Error("expected non-empty buildTime")
	}
}

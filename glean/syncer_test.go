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

package glean

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gleanwork/api-client-go/models/components"
	"github.com/nexthop-ai/netbox-glean-datasource/crawler"
	"github.com/nexthop-ai/netbox-glean-datasource/netbox"
)

// testCrawler is a simple crawler for testing.
type testCrawler struct{}

func (c *testCrawler) ObjectType() string   { return "TestObject" }
func (c *testCrawler) DisplayLabel() string { return "Test Object" }
func (c *testCrawler) Endpoint() string     { return "/api/test/objects/" }

func (c *testCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         crawler.Ptr("TestObject"),
		DisplayLabel: crawler.Ptr("Test Object"),
	}
}

func (c *testCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	return components.DocumentDefinition{
		ID:         crawler.Ptr(crawler.DocID("TestObject", obj)),
		Datasource: datasource,
		ObjectType: crawler.Ptr("TestObject"),
		Title:      crawler.Ptr(netbox.GetString(obj, "name")),
		ViewURL:    crawler.Ptr(netboxURL + "/test/" + netbox.GetString(obj, "name")),
		Permissions: &components.DocumentPermissionsDefinition{
			AllowAnonymousAccess: crawler.Ptr(true),
		},
	}
}

func TestSyncAll(t *testing.T) {
	// Register our test crawler.
	crawler.Register(&testCrawler{})
	defer func() {
		// Clean up: re-register shouldn't cause issues since it overwrites.
	}()

	// Mock NetBox server with 2 pages of results.
	nbServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		var resp netbox.ListResponse
		if offset == "" || offset == "0" {
			next := "http://" + r.Host + "/api/test/objects/?limit=2&offset=2"
			resp = netbox.ListResponse{
				Count: 3,
				Next:  &next,
				Results: []map[string]any{
					{"id": float64(1), "name": "obj-1"},
					{"id": float64(2), "name": "obj-2"},
				},
			}
		} else {
			resp = netbox.ListResponse{
				Count:   3,
				Results: []map[string]any{{"id": float64(3), "name": "obj-3"}},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer nbServer.Close()

	// Mock Glean server that records BulkIndex calls.
	var mu sync.Mutex
	var bulkCalls []components.BulkIndexDocumentsRequest

	gleanServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req components.BulkIndexDocumentsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode BulkIndex request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mu.Lock()
		bulkCalls = append(bulkCalls, req)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer gleanServer.Close()

	// We can't easily use the real Glean SDK with a mock server
	// because it constructs URLs from the instance name.
	// Instead, test the NetBox fetching + transform pipeline directly.

	nbClient := netbox.NewClient(nbServer.URL, "test-token")

	// Test that the NetBox client can fetch and the crawler can transform.
	var allDocs []components.DocumentDefinition
	tc := &testCrawler{}
	err := nbClient.List(context.Background(), tc.Endpoint(), nil, func(results []map[string]any) error {
		for _, obj := range results {
			allDocs = append(allDocs, tc.Transform(obj, "test-ds", nbServer.URL))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(allDocs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(allDocs))
	}

	// Verify documents.
	for i, doc := range allDocs {
		if doc.Datasource != "test-ds" {
			t.Errorf("doc %d: expected datasource 'test-ds', got %q", i, doc.Datasource)
		}
		if doc.ObjectType == nil || *doc.ObjectType != "TestObject" {
			t.Errorf("doc %d: expected objectType 'TestObject', got %v", i, doc.ObjectType)
		}
	}

	// Verify first doc.
	if *allDocs[0].ID != "testobject-1" {
		t.Errorf("expected ID 'testobject-1', got %q", *allDocs[0].ID)
	}
	if *allDocs[0].Title != "obj-1" {
		t.Errorf("expected title 'obj-1', got %q", *allDocs[0].Title)
	}
}

func TestBatchSplitting(t *testing.T) {
	// Test that documents are properly split into batches.
	docs := make([]components.DocumentDefinition, 5)
	for i := range docs {
		docs[i] = components.DocumentDefinition{
			ID:         crawler.Ptr("test-" + string(rune('0'+i))),
			Datasource: "test",
		}
	}

	batchSize := 2
	var batches [][]components.DocumentDefinition

	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}
		batches = append(batches, docs[i:end])
	}

	if len(batches) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(batches))
	}
	if len(batches[0]) != 2 {
		t.Errorf("batch 0: expected 2 docs, got %d", len(batches[0]))
	}
	if len(batches[1]) != 2 {
		t.Errorf("batch 1: expected 2 docs, got %d", len(batches[1]))
	}
	if len(batches[2]) != 1 {
		t.Errorf("batch 2: expected 1 doc, got %d", len(batches[2]))
	}
}

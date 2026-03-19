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

package crawler

import (
	"strings"
	"testing"
)

func TestConsolePortCrawler(t *testing.T) {
	c := &ConsolePortCrawler{}

	if c.ObjectType() != "ConsolePort" {
		t.Errorf("expected ObjectType 'ConsolePort', got %q", c.ObjectType())
	}
	if c.DisplayLabel() != "Console Port" {
		t.Errorf("expected DisplayLabel 'Console Port', got %q", c.DisplayLabel())
	}
	if c.Endpoint() != "/api/dcim/console-ports/" {
		t.Errorf("expected Endpoint '/api/dcim/console-ports/', got %q", c.Endpoint())
	}

	// Test Transform with a sample console port object
	obj := map[string]any{
		"id":      float64(123),
		"display": "Console",
		"name":    "Console",
		"label":   "Console Port 1",
		"device": map[string]any{
			"id":      float64(456),
			"display": "humm210",
		},
		"type": map[string]any{
			"value": "rj-45",
			"label": "RJ-45",
		},
		"speed": map[string]any{
			"value": float64(115200),
			"label": "115.2 kbps",
		},
		"description": "Management console port",
		"cable": map[string]any{
			"id":      float64(789),
			"display": "Cable #789",
			"label":   "CAT6-001",
		},
		"connected_endpoints": []any{
			map[string]any{
				"device": map[string]any{
					"display": "console-server-01",
				},
				"display": "Port 1",
			},
		},
		"mark_connected": true,
		"display_url":    "/dcim/console-ports/123/",
		"created":        "2024-01-01T00:00:00Z",
		"last_updated":   "2024-01-02T00:00:00Z",
	}

	doc := c.Transform(obj, "netbox", "https://netbox.example.com")

	// Check basic fields
	if doc.Datasource != "netbox" {
		t.Errorf("expected datasource 'netbox', got %q", doc.Datasource)
	}
	if doc.ObjectType == nil || *doc.ObjectType != "ConsolePort" {
		t.Errorf("expected objectType 'ConsolePort', got %v", doc.ObjectType)
	}
	if doc.ID == nil || *doc.ID != "consoleport-123" {
		t.Errorf("expected ID 'consoleport-123', got %v", doc.ID)
	}
	if doc.Title == nil || *doc.Title != "Console" {
		t.Errorf("expected title 'Console', got %v", doc.Title)
	}

	// Check container is set to device
	if doc.Container == nil || *doc.Container != "humm210" {
		t.Errorf("expected container 'humm210', got %v", doc.Container)
	}

	// Check body content
	if doc.Body == nil || doc.Body.TextContent == nil {
		t.Fatal("expected body content")
	}
	body := *doc.Body.TextContent
	if !strings.Contains(body, "Name: Console") {
		t.Errorf("body missing 'Name: Console': %s", body)
	}
	if !strings.Contains(body, "Device: humm210") {
		t.Errorf("body missing 'Device: humm210': %s", body)
	}
	if !strings.Contains(body, "Type: RJ-45") {
		t.Errorf("body missing 'Type: RJ-45': %s", body)
	}
	if !strings.Contains(body, "Speed: 115.2 kbps") {
		t.Errorf("body missing 'Speed: 115.2 kbps': %s", body)
	}
	if !strings.Contains(body, "Label: Console Port 1") {
		t.Errorf("body missing 'Label: Console Port 1': %s", body)
	}
	if !strings.Contains(body, "Connected To: console-server-01:Port 1") {
		t.Errorf("body missing 'Connected To: console-server-01:Port 1': %s", body)
	}
	if !strings.Contains(body, "Cable: CAT6-001") {
		t.Errorf("body missing 'Cable: CAT6-001': %s", body)
	}

	// Check custom properties
	propMap := make(map[string]string)
	for _, prop := range doc.CustomProperties {
		if prop.Name != nil {
			propMap[*prop.Name] = prop.Value.(string)
		}
	}

	if propMap["nbDevice"] != "humm210" {
		t.Errorf("expected nbDevice 'humm210', got %q", propMap["nbDevice"])
	}
	if propMap["nbPortType"] != "RJ-45" {
		t.Errorf("expected nbPortType 'RJ-45', got %q", propMap["nbPortType"])
	}
	if propMap["nbSpeed"] != "115.2 kbps" {
		t.Errorf("expected nbSpeed '115.2 kbps', got %q", propMap["nbSpeed"])
	}
	if propMap["nbLabel"] != "Console Port 1" {
		t.Errorf("expected nbLabel 'Console Port 1', got %q", propMap["nbLabel"])
	}
	if propMap["nbConnectedTo"] != "console-server-01:Port 1" {
		t.Errorf("expected nbConnectedTo 'console-server-01:Port 1', got %q", propMap["nbConnectedTo"])
	}
	if propMap["nbCable"] != "CAT6-001" {
		t.Errorf("expected nbCable 'CAT6-001', got %q", propMap["nbCable"])
	}
}

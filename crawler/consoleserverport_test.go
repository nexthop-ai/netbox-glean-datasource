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
	"testing"
)

func TestConsoleServerPortCrawler(t *testing.T) {
	c := &ConsoleServerPortCrawler{}

	if c.ObjectType() != "ConsoleServerPort" {
		t.Errorf("expected ObjectType 'ConsoleServerPort', got %q", c.ObjectType())
	}
	if c.DisplayLabel() != "Console Server Port" {
		t.Errorf("expected DisplayLabel 'Console Server Port', got %q", c.DisplayLabel())
	}
	if c.Endpoint() != "/api/dcim/console-server-ports/" {
		t.Errorf("expected Endpoint '/api/dcim/console-server-ports/', got %q", c.Endpoint())
	}

	// Test Transform with a sample console server port object
	obj := map[string]any{
		"id":      float64(456),
		"display": "Port 1",
		"name":    "Port 1",
		"label":   "Console Server Port 1",
		"device": map[string]any{
			"id":      float64(789),
			"display": "console-server-01",
		},
		"type": map[string]any{
			"value": "rj-45",
			"label": "RJ-45",
		},
		"speed": map[string]any{
			"value": float64(115200),
			"label": "115.2 kbps",
		},
		"description": "Console server port for remote access",
		"cable": map[string]any{
			"id":      float64(789),
			"display": "Cable #789",
			"label":   "CAT6-001",
		},
		"connected_endpoints": []any{
			map[string]any{
				"device": map[string]any{
					"display": "humm210",
				},
				"display": "Console",
			},
		},
		"mark_connected": true,
		"display_url":    "/dcim/console-server-ports/456/",
		"created":        "2024-01-01T00:00:00Z",
		"last_updated":   "2024-01-02T00:00:00Z",
	}

	doc := c.Transform(obj, "netbox", "https://netbox.example.com")

	// Check basic fields
	if doc.Datasource != "netbox" {
		t.Errorf("expected datasource 'netbox', got %q", doc.Datasource)
	}
	if doc.ObjectType == nil || *doc.ObjectType != "ConsoleServerPort" {
		t.Errorf("expected objectType 'ConsoleServerPort', got %v", doc.ObjectType)
	}
	if doc.ID == nil || *doc.ID != "consoleserverport-456" {
		t.Errorf("expected ID 'consoleserverport-456', got %v", doc.ID)
	}
	if doc.Title == nil || *doc.Title != "Port 1" {
		t.Errorf("expected title 'Port 1', got %v", doc.Title)
	}

	// Check container (parent device)
	if doc.Container == nil || *doc.Container != "console-server-01" {
		t.Errorf("expected container 'console-server-01', got %v", doc.Container)
	}

	// Check custom properties
	props := doc.CustomProperties
	if len(props) == 0 {
		t.Fatal("expected custom properties, got none")
	}

	// Verify device property
	foundDevice := false
	for _, p := range props {
		if p.Name != nil && *p.Name == "nbDevice" && p.Value == "console-server-01" {
			foundDevice = true
			break
		}
	}
	if !foundDevice {
		t.Error("expected nbDevice property with value 'console-server-01'")
	}

	// Verify port type property
	foundType := false
	for _, p := range props {
		if p.Name != nil && *p.Name == "nbPortType" && p.Value == "RJ-45" {
			foundType = true
			break
		}
	}
	if !foundType {
		t.Error("expected nbPortType property with value 'RJ-45'")
	}

	// Verify speed property
	foundSpeed := false
	for _, p := range props {
		if p.Name != nil && *p.Name == "nbSpeed" && p.Value == "115.2 kbps" {
			foundSpeed = true
			break
		}
	}
	if !foundSpeed {
		t.Error("expected nbSpeed property with value '115.2 kbps'")
	}

	// Verify connected to property
	foundConnectedTo := false
	for _, p := range props {
		if p.Name != nil && *p.Name == "nbConnectedTo" {
			foundConnectedTo = true
			// Should contain the connected device name
			if p.Value != "humm210:Console" {
				t.Errorf("expected nbConnectedTo to contain 'humm210:Console', got %q", p.Value)
			}
			break
		}
	}
	if !foundConnectedTo {
		t.Error("expected nbConnectedTo property")
	}
}

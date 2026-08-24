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

func TestPowerPortCrawler(t *testing.T) {
	c := &PowerPortCrawler{}

	if c.ObjectType() != "PowerPort" {
		t.Errorf("expected ObjectType 'PowerPort', got %q", c.ObjectType())
	}
	if c.DisplayLabel() != "Power Port" {
		t.Errorf("expected DisplayLabel 'Power Port', got %q", c.DisplayLabel())
	}
	if c.Endpoint() != "/api/dcim/power-ports/" {
		t.Errorf("expected Endpoint '/api/dcim/power-ports/', got %q", c.Endpoint())
	}

	// Test Transform with a sample power port (PSU) object
	obj := map[string]any{
		"id":      float64(321),
		"display": "PSU1",
		"name":    "PSU1",
		"label":   "Power Supply 1",
		"device": map[string]any{
			"id":      float64(456),
			"display": "humm210",
		},
		"type": map[string]any{
			"value": "iec-60320-c14",
			"label": "C14",
		},
		"maximum_draw":   float64(750),
		"allocated_draw": float64(500),
		"description":    "Primary power supply",
		"cable": map[string]any{
			"id":      float64(987),
			"display": "Cable #987",
			"label":   "PWR-001",
		},
		"connected_endpoints": []any{
			map[string]any{
				"device": map[string]any{
					"display": "pdu-a-01",
				},
				"display": "Outlet 4",
			},
		},
		"mark_connected": true,
		"display_url":    "/dcim/power-ports/321/",
		"created":        "2024-01-01T00:00:00Z",
		"last_updated":   "2024-01-02T00:00:00Z",
	}

	doc := c.Transform(obj, "netbox", "https://netbox.example.com")

	if doc.Datasource != "netbox" {
		t.Errorf("expected datasource 'netbox', got %q", doc.Datasource)
	}
	if doc.ObjectType == nil || *doc.ObjectType != "PowerPort" {
		t.Errorf("expected objectType 'PowerPort', got %v", doc.ObjectType)
	}
	if doc.ID == nil || *doc.ID != "powerport-321" {
		t.Errorf("expected ID 'powerport-321', got %v", doc.ID)
	}
	if doc.Container == nil || *doc.Container != "humm210" {
		t.Errorf("expected container 'humm210', got %v", doc.Container)
	}

	if doc.Body == nil || doc.Body.TextContent == nil {
		t.Fatal("expected body content")
	}
	body := *doc.Body.TextContent
	for _, want := range []string{
		"Name: PSU1",
		"Device: humm210",
		"Type: C14",
		"Label: Power Supply 1",
		"Maximum Draw (W): 750",
		"Allocated Draw (W): 500",
		"Connected To: pdu-a-01:Outlet 4",
		"Cable: PWR-001",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q: %s", want, body)
		}
	}

	propMap := make(map[string]string)
	for _, prop := range doc.CustomProperties {
		if prop.Name != nil {
			propMap[*prop.Name] = prop.Value.(string)
		}
	}

	if propMap["nbDevice"] != "humm210" {
		t.Errorf("expected nbDevice 'humm210', got %q", propMap["nbDevice"])
	}
	if propMap["nbPortType"] != "C14" {
		t.Errorf("expected nbPortType 'C14', got %q", propMap["nbPortType"])
	}
	if propMap["nbLabel"] != "Power Supply 1" {
		t.Errorf("expected nbLabel 'Power Supply 1', got %q", propMap["nbLabel"])
	}
	if propMap["nbMaximumDraw"] != "750" {
		t.Errorf("expected nbMaximumDraw '750', got %q", propMap["nbMaximumDraw"])
	}
	if propMap["nbAllocatedDraw"] != "500" {
		t.Errorf("expected nbAllocatedDraw '500', got %q", propMap["nbAllocatedDraw"])
	}
	if propMap["nbConnectedTo"] != "pdu-a-01:Outlet 4" {
		t.Errorf("expected nbConnectedTo 'pdu-a-01:Outlet 4', got %q", propMap["nbConnectedTo"])
	}
	if propMap["nbCable"] != "PWR-001" {
		t.Errorf("expected nbCable 'PWR-001', got %q", propMap["nbCable"])
	}
}

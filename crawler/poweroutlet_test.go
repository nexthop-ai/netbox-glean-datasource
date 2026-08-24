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

func TestPowerOutletCrawler(t *testing.T) {
	c := &PowerOutletCrawler{}

	if c.ObjectType() != "PowerOutlet" {
		t.Errorf("expected ObjectType 'PowerOutlet', got %q", c.ObjectType())
	}
	if c.DisplayLabel() != "Power Outlet" {
		t.Errorf("expected DisplayLabel 'Power Outlet', got %q", c.DisplayLabel())
	}
	if c.Endpoint() != "/api/dcim/power-outlets/" {
		t.Errorf("expected Endpoint '/api/dcim/power-outlets/', got %q", c.Endpoint())
	}

	// Test Transform with a sample PDU power outlet object
	obj := map[string]any{
		"id":      float64(654),
		"display": "Outlet 4",
		"name":    "Outlet 4",
		"label":   "A4",
		"device": map[string]any{
			"id":      float64(111),
			"display": "pdu-a-01",
		},
		"type": map[string]any{
			"value": "iec-60320-c13",
			"label": "C13",
		},
		"power_port": map[string]any{
			"id":      float64(222),
			"display": "Feed A",
		},
		"feed_leg": map[string]any{
			"value": "A",
			"label": "A",
		},
		"description": "Rack A outlet 4",
		"cable": map[string]any{
			"id":      float64(987),
			"display": "Cable #987",
			"label":   "PWR-001",
		},
		"connected_endpoints": []any{
			map[string]any{
				"device": map[string]any{
					"display": "humm210",
				},
				"display": "PSU1",
			},
		},
		"mark_connected": true,
		"display_url":    "/dcim/power-outlets/654/",
		"created":        "2024-01-01T00:00:00Z",
		"last_updated":   "2024-01-02T00:00:00Z",
	}

	doc := c.Transform(obj, "netbox", "https://netbox.example.com")

	if doc.Datasource != "netbox" {
		t.Errorf("expected datasource 'netbox', got %q", doc.Datasource)
	}
	if doc.ObjectType == nil || *doc.ObjectType != "PowerOutlet" {
		t.Errorf("expected objectType 'PowerOutlet', got %v", doc.ObjectType)
	}
	if doc.ID == nil || *doc.ID != "poweroutlet-654" {
		t.Errorf("expected ID 'poweroutlet-654', got %v", doc.ID)
	}
	if doc.Container == nil || *doc.Container != "pdu-a-01" {
		t.Errorf("expected container 'pdu-a-01', got %v", doc.Container)
	}

	if doc.Body == nil || doc.Body.TextContent == nil {
		t.Fatal("expected body content")
	}
	body := *doc.Body.TextContent
	for _, want := range []string{
		"Name: Outlet 4",
		"Device: pdu-a-01",
		"Type: C13",
		"Label: A4",
		"Power Port: Feed A",
		"Feed Leg: A",
		"Connected To: humm210:PSU1",
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

	if propMap["nbDevice"] != "pdu-a-01" {
		t.Errorf("expected nbDevice 'pdu-a-01', got %q", propMap["nbDevice"])
	}
	if propMap["nbPortType"] != "C13" {
		t.Errorf("expected nbPortType 'C13', got %q", propMap["nbPortType"])
	}
	if propMap["nbLabel"] != "A4" {
		t.Errorf("expected nbLabel 'A4', got %q", propMap["nbLabel"])
	}
	if propMap["nbPowerPort"] != "Feed A" {
		t.Errorf("expected nbPowerPort 'Feed A', got %q", propMap["nbPowerPort"])
	}
	if propMap["nbFeedLeg"] != "A" {
		t.Errorf("expected nbFeedLeg 'A', got %q", propMap["nbFeedLeg"])
	}
	if propMap["nbConnectedTo"] != "humm210:PSU1" {
		t.Errorf("expected nbConnectedTo 'humm210:PSU1', got %q", propMap["nbConnectedTo"])
	}
	if propMap["nbCable"] != "PWR-001" {
		t.Errorf("expected nbCable 'PWR-001', got %q", propMap["nbCable"])
	}
}

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

func TestDeviceTransform(t *testing.T) {
	obj := map[string]any{
		"id":      float64(42),
		"display": "core-sw-01",
		"display_url": "/dcim/devices/42/",
		"status": map[string]any{
			"value": "active",
			"label": "Active",
		},
		"site": map[string]any{
			"display": "NYC-DC1",
		},
		"location": map[string]any{
			"display": "Floor 3",
		},
		"rack": map[string]any{
			"display": "Rack A1",
		},
		"position": float64(20),
		"role": map[string]any{
			"display": "Core Switch",
		},
		"device_type": map[string]any{
			"display": "Catalyst 9300",
			"manufacturer": map[string]any{
				"display": "Cisco",
			},
		},
		"platform": map[string]any{
			"display": "IOS-XE",
		},
		"tenant": map[string]any{
			"display": "Engineering",
		},
		"serial":    "FOC1234ABCD",
		"asset_tag": "ASSET-001",
		"primary_ip4": map[string]any{
			"display": "10.0.1.1/24",
		},
		"primary_ip6": nil,
		"oob_ip":      nil,
		"description": "Main core switch for NYC datacenter",
		"tags": []any{
			map[string]any{"display": "production", "name": "production"},
			map[string]any{"display": "core", "name": "core"},
		},
		"created":      "2024-01-15T10:00:00Z",
		"last_updated": "2024-06-15T10:30:00Z",
	}

	c := &DeviceCrawler{}
	doc := c.Transform(obj, "netbox", "https://netbox.example.com")

	// Check ID.
	if doc.ID == nil || *doc.ID != "device-42" {
		t.Errorf("expected ID 'device-42', got %v", doc.ID)
	}

	// Check title.
	if doc.Title == nil || *doc.Title != "core-sw-01" {
		t.Errorf("expected title 'core-sw-01', got %v", doc.Title)
	}

	// Check ViewURL.
	if doc.ViewURL == nil || *doc.ViewURL != "https://netbox.example.com/dcim/devices/42/" {
		t.Errorf("expected ViewURL 'https://netbox.example.com/dcim/devices/42/', got %v", doc.ViewURL)
	}

	// Check datasource.
	if doc.Datasource != "netbox" {
		t.Errorf("expected datasource 'netbox', got %q", doc.Datasource)
	}

	// Check object type.
	if doc.ObjectType == nil || *doc.ObjectType != "Device" {
		t.Errorf("expected objectType 'Device', got %v", doc.ObjectType)
	}

	// Check body contains key fields.
	if doc.Body == nil || doc.Body.TextContent == nil {
		t.Fatal("expected non-nil body")
	}
	body := *doc.Body.TextContent
	for _, expected := range []string{
		"core-sw-01", "Active", "NYC-DC1", "Core Switch",
		"Catalyst 9300", "Cisco", "IOS-XE", "Engineering",
		"FOC1234ABCD", "ASSET-001", "10.0.1.1/24",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("body missing %q", expected)
		}
	}

	// Check tags.
	if len(doc.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(doc.Tags))
	}

	// Check status.
	if doc.Status == nil || *doc.Status != "Active" {
		t.Errorf("expected status 'Active', got %v", doc.Status)
	}

	// Check permissions.
	if doc.Permissions == nil || doc.Permissions.AllowAnonymousAccess == nil || !*doc.Permissions.AllowAnonymousAccess {
		t.Error("expected AllowAnonymousAccess = true")
	}

	// Check timestamps.
	if doc.CreatedAt == nil {
		t.Error("expected non-nil CreatedAt")
	}
	if doc.UpdatedAt == nil {
		t.Error("expected non-nil UpdatedAt")
	}

	// Check custom properties.
	if len(doc.CustomProperties) == 0 {
		t.Fatal("expected custom properties")
	}
	propMap := make(map[string]string)
	for _, p := range doc.CustomProperties {
		if p.Name != nil {
			propMap[*p.Name] = p.Value.(string)
		}
	}
	expectedProps := map[string]string{
		"nbStatus":       "Active",
		"nbSite":         "NYC-DC1",
		"nbRole":         "Core Switch",
		"nbTenant":       "Engineering",
		"nbDeviceType":   "Catalyst 9300",
		"nbPlatform":     "IOS-XE",
		"nbRack":         "Rack A1",
		"nbSerialNumber": "FOC1234ABCD",
		"nbAssetTag":     "ASSET-001",
		"nbPrimaryIP":    "10.0.1.1/24",
	}
	for key, expected := range expectedProps {
		if got, ok := propMap[key]; !ok {
			t.Errorf("missing custom property %q", key)
		} else if got != expected {
			t.Errorf("property %q: expected %q, got %q", key, expected, got)
		}
	}
}

func TestDeviceTransformMinimal(t *testing.T) {
	obj := map[string]any{
		"id":      float64(1),
		"display": "unnamed-device",
	}

	c := &DeviceCrawler{}
	doc := c.Transform(obj, "netbox", "https://netbox.example.com")

	if doc.ID == nil || *doc.ID != "device-1" {
		t.Errorf("expected ID 'device-1', got %v", doc.ID)
	}
	if doc.Title == nil || *doc.Title != "unnamed-device" {
		t.Errorf("expected title 'unnamed-device', got %v", doc.Title)
	}
	if doc.Body == nil || doc.Body.TextContent == nil {
		t.Fatal("expected non-nil body even for minimal object")
	}
}

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

func TestRegistry(t *testing.T) {
	// All crawlers register in init(), so the registry should be populated.
	all := All()
	if len(all) == 0 {
		t.Fatal("expected registered crawlers, got none")
	}

	// Check that we have the expected types.
	expectedTypes := []string{
		"Device", "Site", "IPAddress", "Prefix", "VLAN",
		"VirtualMachine", "Rack", "Circuit", "Tenant", "VRF",
		"Cluster", "Interface", "Provider", "Platform", "Manufacturer",
	}
	for _, ot := range expectedTypes {
		c, ok := Get(ot)
		if !ok {
			t.Errorf("expected crawler for %q, not found", ot)
			continue
		}
		if c.ObjectType() != ot {
			t.Errorf("expected ObjectType() = %q, got %q", ot, c.ObjectType())
		}
		if c.DisplayLabel() == "" {
			t.Errorf("empty DisplayLabel for %q", ot)
		}
		if c.Endpoint() == "" {
			t.Errorf("empty Endpoint for %q", ot)
		}
		od := c.ObjectDefinition()
		if od.Name == nil || *od.Name != ot {
			t.Errorf("expected ObjectDefinition.Name = %q, got %v", ot, od.Name)
		}
	}

	if len(all) != len(expectedTypes) {
		t.Errorf("expected %d crawlers, got %d", len(expectedTypes), len(all))
	}
}

func TestPtr(t *testing.T) {
	s := Ptr("hello")
	if *s != "hello" {
		t.Errorf("expected 'hello', got %q", *s)
	}

	i := Ptr(42)
	if *i != 42 {
		t.Errorf("expected 42, got %d", *i)
	}

	b := Ptr(true)
	if !*b {
		t.Error("expected true")
	}
}

func TestBuildViewURL(t *testing.T) {
	tests := []struct {
		name     string
		obj      map[string]any
		base     string
		expected string
	}{
		{
			name:     "with display_url",
			obj:      map[string]any{"display_url": "/dcim/devices/42/"},
			base:     "https://netbox.example.com",
			expected: "https://netbox.example.com/dcim/devices/42/",
		},
		{
			name:     "with display_url trailing slash base",
			obj:      map[string]any{"display_url": "/dcim/devices/42/"},
			base:     "https://netbox.example.com/",
			expected: "https://netbox.example.com/dcim/devices/42/",
		},
		{
			name:     "with absolute display_url",
			obj:      map[string]any{"display_url": "https://netbox.example.com/dcim/devices/42/"},
			base:     "https://netbox.example.com",
			expected: "https://netbox.example.com/dcim/devices/42/",
		},
		{
			name:     "fallback to url",
			obj:      map[string]any{"url": "https://netbox.example.com/api/dcim/devices/42/"},
			base:     "https://netbox.example.com",
			expected: "https://netbox.example.com/api/dcim/devices/42/",
		},
		{
			name:     "fallback to base",
			obj:      map[string]any{},
			base:     "https://netbox.example.com",
			expected: "https://netbox.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildViewURL(tt.obj, tt.base)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEpochSeconds(t *testing.T) {
	obj := map[string]any{
		"created":      "2024-06-15T10:30:00Z",
		"last_updated": "2024-06-15T10:30:00Z",
	}

	epoch := EpochSeconds(obj, "created")
	if epoch == nil {
		t.Fatal("expected non-nil epoch")
	}
	// 2024-06-15T10:30:00Z
	if *epoch != 1718447400 {
		t.Errorf("expected 1718447400, got %d", *epoch)
	}

	missing := EpochSeconds(obj, "nonexistent")
	if missing != nil {
		t.Errorf("expected nil for missing key, got %d", *missing)
	}
}

func TestDocID(t *testing.T) {
	obj := map[string]any{"id": float64(42)}
	id := DocID("Device", obj)
	if id != "device-42" {
		t.Errorf("expected 'device-42', got %q", id)
	}
}

func TestStatusValue(t *testing.T) {
	// Object-style status (NetBox default).
	obj1 := map[string]any{
		"status": map[string]any{"value": "active", "label": "Active"},
	}
	if v := StatusValue(obj1); v != "Active" {
		t.Errorf("expected 'Active', got %q", v)
	}

	// String status.
	obj2 := map[string]any{"status": "active"}
	if v := StatusValue(obj2); v != "active" {
		t.Errorf("expected 'active', got %q", v)
	}

	// No status.
	obj3 := map[string]any{}
	if v := StatusValue(obj3); v != "" {
		t.Errorf("expected empty, got %q", v)
	}
}

func TestBodyBuilder(t *testing.T) {
	var bb BodyBuilder
	bb.Add("Name", "test-device")
	bb.Add("Status", "Active")
	bb.Add("Empty", "")

	result := bb.String()
	expected := "Name: test-device\nStatus: Active\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCustomProp(t *testing.T) {
	p := CustomProp("nbSite", "Site A")
	if *p.Name != "nbSite" {
		t.Errorf("expected name 'nbSite', got %q", *p.Name)
	}
	if p.Value != "Site A" {
		t.Errorf("expected value 'Site A', got %v", p.Value)
	}
}

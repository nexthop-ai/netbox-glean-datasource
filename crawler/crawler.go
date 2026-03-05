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
	"fmt"
	"strings"
	"time"

	"github.com/gleanwork/api-client-go/models/components"
	"github.com/nexthop-ai/netbox-glean-datasource/netbox"
)

// Crawler defines the interface for each NetBox object type.
type Crawler interface {
	// ObjectType returns the Glean objectType string (e.g., "Device").
	ObjectType() string
	// DisplayLabel returns the human-readable label for this object type.
	DisplayLabel() string
	// Endpoint returns the NetBox API endpoint path (e.g., "/api/dcim/devices/").
	Endpoint() string
	// ObjectDefinition returns the Glean schema for this object type.
	ObjectDefinition() components.ObjectDefinition
	// Transform converts a single NetBox API result into a Glean DocumentDefinition.
	Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition
}

var registry = map[string]Crawler{}

func Register(c Crawler) {
	registry[c.ObjectType()] = c
}

func All() map[string]Crawler {
	return registry
}

func Get(objectType string) (Crawler, bool) {
	c, ok := registry[objectType]
	return c, ok
}

// Shared helpers used by all crawlers.

func Ptr[T any](v T) *T {
	return &v
}

func BuildViewURL(obj map[string]any, netboxURL string) string {
	if u := netbox.GetString(obj, "display_url"); u != "" {
		if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
			return u
		}
		return strings.TrimRight(netboxURL, "/") + u
	}
	if u := netbox.GetString(obj, "url"); u != "" {
		return u
	}
	return netboxURL
}

func EpochSeconds(obj map[string]any, key string) *int64 {
	t := netbox.GetTime(obj, key)
	if t == nil {
		return nil
	}
	epoch := t.Unix()
	return &epoch
}

func DocID(objectType string, obj map[string]any) string {
	id := netbox.GetInt(obj, "id")
	return fmt.Sprintf("%s-%d", strings.ToLower(objectType), id)
}

func CustomProp(name, value string) components.CustomProperty {
	return components.CustomProperty{
		Name:  Ptr(name),
		Value: value,
	}
}

func CustomProps(obj map[string]any, fields ...string) []components.CustomProperty {
	var props []components.CustomProperty
	for _, field := range fields {
		v := netbox.GetString(obj, field)
		if v != "" {
			props = append(props, CustomProp(field, v))
		}
	}
	return props
}

func NestedCustomProp(name string, obj map[string]any, keys ...string) *components.CustomProperty {
	v := netbox.GetNestedString(obj, keys...)
	if v == "" {
		return nil
	}
	p := CustomProp(name, v)
	return &p
}

func StatusValue(obj map[string]any) string {
	// NetBox status fields are objects with {value, label} or plain strings.
	v, ok := obj["status"]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case map[string]any:
		return netbox.GetString(s, "label")
	case string:
		return s
	default:
		return ""
	}
}

// BodyBuilder helps construct the plain text body for full-text search.
type BodyBuilder struct {
	b strings.Builder
}

func (bb *BodyBuilder) Add(label, value string) {
	if value == "" {
		return
	}
	bb.b.WriteString(label)
	bb.b.WriteString(": ")
	bb.b.WriteString(value)
	bb.b.WriteString("\n")
}

func (bb *BodyBuilder) AddNested(label string, obj map[string]any, keys ...string) {
	bb.Add(label, netbox.GetNestedString(obj, keys...))
}

func (bb *BodyBuilder) String() string {
	return bb.b.String()
}

func BaseDocument(objectType string, obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	return components.DocumentDefinition{
		ID:         Ptr(DocID(objectType, obj)),
		Datasource: datasource,
		ObjectType: Ptr(objectType),
		Title:      Ptr(netbox.GetString(obj, "display")),
		ViewURL:    Ptr(BuildViewURL(obj, netboxURL)),
		Permissions: &components.DocumentPermissionsDefinition{
			AllowAnonymousAccess: Ptr(true),
		},
		CreatedAt: EpochSeconds(obj, "created"),
		UpdatedAt: EpochSeconds(obj, "last_updated"),
		Tags:      netbox.GetTags(obj),
	}
}

func PropertyDef(name, label string, propType components.PropertyType) components.PropertyDefinition {
	return components.PropertyDefinition{
		Name:         Ptr(name),
		DisplayLabel: Ptr(label),
		PropertyType: propType.ToPointer(),
		HideUIFacet:  Ptr(true),
	}
}

func FacetDef(name, label string, propType components.PropertyType, order int64) components.PropertyDefinition {
	p := PropertyDef(name, label, propType)
	p.UIOptions = components.UIOptionsSearchResult.ToPointer()
	p.UIFacetOrder = &order
	return p
}

// DescriptionBody returns a ContentDefinition with the object's description as fallback.
func DescriptionBody(obj map[string]any) *components.ContentDefinition {
	desc := netbox.GetString(obj, "description")
	if desc == "" {
		return nil
	}
	return &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(desc),
	}
}

// FormatDuration formats a Go duration to a human-readable string.
func FormatDuration(d time.Duration) string {
	return d.String()
}

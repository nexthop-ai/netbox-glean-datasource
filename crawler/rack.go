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

	"github.com/gleanwork/api-client-go/models/components"
	"github.com/nexthop-ai/netbox-glean-datasource/netbox"
)

type RackCrawler struct{}

func init() { Register(&RackCrawler{}) }

func (c *RackCrawler) ObjectType() string   { return "Rack" }
func (c *RackCrawler) DisplayLabel() string { return "Rack" }
func (c *RackCrawler) Endpoint() string     { return "/api/dcim/racks/" }

func (c *RackCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("Rack"),
		DisplayLabel: Ptr("Rack"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbStatus", "Status", components.PropertyTypePicklist, 1),
			FacetDef("nbSite", "Site", components.PropertyTypePicklist, 2),
			FacetDef("nbRole", "Role", components.PropertyTypePicklist, 3),
			FacetDef("nbTenant", "Tenant", components.PropertyTypePicklist, 4),
			PropertyDef("nbLocation", "Location", components.PropertyTypeText),
			PropertyDef("nbUHeight", "U Height", components.PropertyTypeText),
			PropertyDef("nbSerialNumber", "Serial Number", components.PropertyTypeText),
		},
	}
}

func (c *RackCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("Rack", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.Add("Status", StatusValue(obj))
	bb.AddNested("Site", obj, "site", "display")
	bb.AddNested("Location", obj, "location", "display")
	bb.AddNested("Role", obj, "role", "display")
	bb.AddNested("Tenant", obj, "tenant", "display")
	if v := netbox.GetInt(obj, "u_height"); v > 0 {
		bb.Add("U Height", fmt.Sprintf("%d", v))
	}
	bb.Add("Serial", netbox.GetString(obj, "serial"))
	bb.Add("Asset Tag", netbox.GetString(obj, "asset_tag"))
	bb.Add("Facility ID", netbox.GetString(obj, "facility_id"))
	bb.Add("Description", netbox.GetString(obj, "description"))

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}
	doc.Status = Ptr(StatusValue(obj))

	var props []components.CustomProperty
	if v := StatusValue(obj); v != "" {
		props = append(props, CustomProp("nbStatus", v))
	}
	if v := netbox.GetNestedString(obj, "site", "display"); v != "" {
		props = append(props, CustomProp("nbSite", v))
	}
	if v := netbox.GetNestedString(obj, "role", "display"); v != "" {
		props = append(props, CustomProp("nbRole", v))
	}
	if v := netbox.GetNestedString(obj, "tenant", "display"); v != "" {
		props = append(props, CustomProp("nbTenant", v))
	}
	if v := netbox.GetNestedString(obj, "location", "display"); v != "" {
		props = append(props, CustomProp("nbLocation", v))
	}
	if v := netbox.GetInt(obj, "u_height"); v > 0 {
		props = append(props, CustomProp("nbUHeight", fmt.Sprintf("%d", v)))
	}
	if v := netbox.GetString(obj, "serial"); v != "" {
		props = append(props, CustomProp("nbSerialNumber", v))
	}
	doc.CustomProperties = props

	return doc
}

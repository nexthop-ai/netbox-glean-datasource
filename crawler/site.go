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
	"github.com/gleanwork/api-client-go/models/components"
	"github.com/nexthop-ai/netbox-glean-datasource/netbox"
)

type SiteCrawler struct{}

func init() { Register(&SiteCrawler{}) }

func (s *SiteCrawler) ObjectType() string   { return "Site" }
func (s *SiteCrawler) DisplayLabel() string { return "Site" }
func (s *SiteCrawler) Endpoint() string     { return "/api/dcim/sites/" }

func (s *SiteCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("Site"),
		DisplayLabel: Ptr("Site"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbStatus", "Status", components.PropertyTypePicklist, 1),
			FacetDef("nbRegion", "Region", components.PropertyTypePicklist, 2),
			FacetDef("nbGroup", "Group", components.PropertyTypePicklist, 3),
			FacetDef("nbTenant", "Tenant", components.PropertyTypePicklist, 4),
			PropertyDef("nbFacility", "Facility", components.PropertyTypeText),
			PropertyDef("nbPhysicalAddress", "Physical Address", components.PropertyTypeText),
			PropertyDef("nbTimezone", "Time Zone", components.PropertyTypeText),
		},
	}
}

func (s *SiteCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("Site", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.Add("Status", StatusValue(obj))
	bb.AddNested("Region", obj, "region", "display")
	bb.AddNested("Group", obj, "group", "display")
	bb.AddNested("Tenant", obj, "tenant", "display")
	bb.Add("Facility", netbox.GetString(obj, "facility"))
	bb.Add("Time Zone", netbox.GetString(obj, "time_zone"))
	bb.Add("Physical Address", netbox.GetString(obj, "physical_address"))
	bb.Add("Shipping Address", netbox.GetString(obj, "shipping_address"))
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
	if v := netbox.GetNestedString(obj, "region", "display"); v != "" {
		props = append(props, CustomProp("nbRegion", v))
	}
	if v := netbox.GetNestedString(obj, "group", "display"); v != "" {
		props = append(props, CustomProp("nbGroup", v))
	}
	if v := netbox.GetNestedString(obj, "tenant", "display"); v != "" {
		props = append(props, CustomProp("nbTenant", v))
	}
	if v := netbox.GetString(obj, "facility"); v != "" {
		props = append(props, CustomProp("nbFacility", v))
	}
	if v := netbox.GetString(obj, "physical_address"); v != "" {
		props = append(props, CustomProp("nbPhysicalAddress", v))
	}
	if v := netbox.GetString(obj, "time_zone"); v != "" {
		props = append(props, CustomProp("nbTimezone", v))
	}
	doc.CustomProperties = props

	return doc
}

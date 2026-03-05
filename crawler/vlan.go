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

type VLANCrawler struct{}

func init() { Register(&VLANCrawler{}) }

func (c *VLANCrawler) ObjectType() string   { return "VLAN" }
func (c *VLANCrawler) DisplayLabel() string { return "VLAN" }
func (c *VLANCrawler) Endpoint() string     { return "/api/ipam/vlans/" }

func (c *VLANCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("VLAN"),
		DisplayLabel: Ptr("VLAN"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbStatus", "Status", components.PropertyTypePicklist, 1),
			FacetDef("nbRole", "Role", components.PropertyTypePicklist, 2),
			FacetDef("nbSite", "Site", components.PropertyTypePicklist, 3),
			FacetDef("nbTenant", "Tenant", components.PropertyTypePicklist, 4),
			PropertyDef("nbVid", "VID", components.PropertyTypeText),
			PropertyDef("nbGroup", "Group", components.PropertyTypeText),
		},
	}
}

func (c *VLANCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("VLAN", obj, datasource, netboxURL)

	vid := netbox.GetInt(obj, "vid")
	var bb BodyBuilder
	bb.Add("VLAN", netbox.GetString(obj, "display"))
	bb.Add("VID", fmt.Sprintf("%d", vid))
	bb.Add("Name", netbox.GetString(obj, "name"))
	bb.Add("Status", StatusValue(obj))
	bb.AddNested("Site", obj, "site", "display")
	bb.AddNested("Group", obj, "group", "display")
	bb.AddNested("Tenant", obj, "tenant", "display")
	bb.AddNested("Role", obj, "role", "display")
	bb.Add("Description", netbox.GetString(obj, "description"))

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}
	doc.Status = Ptr(StatusValue(obj))

	var props []components.CustomProperty
	if vid > 0 {
		props = append(props, CustomProp("nbVid", fmt.Sprintf("%d", vid)))
	}
	if v := StatusValue(obj); v != "" {
		props = append(props, CustomProp("nbStatus", v))
	}
	if v := netbox.GetNestedString(obj, "role", "display"); v != "" {
		props = append(props, CustomProp("nbRole", v))
	}
	if v := netbox.GetNestedString(obj, "site", "display"); v != "" {
		props = append(props, CustomProp("nbSite", v))
	}
	if v := netbox.GetNestedString(obj, "tenant", "display"); v != "" {
		props = append(props, CustomProp("nbTenant", v))
	}
	if v := netbox.GetNestedString(obj, "group", "display"); v != "" {
		props = append(props, CustomProp("nbGroup", v))
	}
	doc.CustomProperties = props

	return doc
}

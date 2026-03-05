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

type PrefixCrawler struct{}

func init() { Register(&PrefixCrawler{}) }

func (c *PrefixCrawler) ObjectType() string   { return "Prefix" }
func (c *PrefixCrawler) DisplayLabel() string { return "Prefix" }
func (c *PrefixCrawler) Endpoint() string     { return "/api/ipam/prefixes/" }

func (c *PrefixCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("Prefix"),
		DisplayLabel: Ptr("Prefix"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbStatus", "Status", components.PropertyTypePicklist, 1),
			FacetDef("nbRole", "Role", components.PropertyTypePicklist, 2),
			FacetDef("nbVrf", "VRF", components.PropertyTypePicklist, 3),
			FacetDef("nbTenant", "Tenant", components.PropertyTypePicklist, 4),
			PropertyDef("nbVlan", "VLAN", components.PropertyTypeText),
			PropertyDef("nbIsPool", "Is Pool", components.PropertyTypeText),
			PropertyDef("nbPrefix", "Prefix", components.PropertyTypeText),
		},
	}
}

func (c *PrefixCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("Prefix", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Prefix", netbox.GetString(obj, "display"))
	bb.Add("Status", StatusValue(obj))
	bb.AddNested("VRF", obj, "vrf", "display")
	bb.AddNested("Tenant", obj, "tenant", "display")
	bb.AddNested("VLAN", obj, "vlan", "display")
	bb.AddNested("Role", obj, "role", "display")
	bb.Add("Is Pool", fmt.Sprintf("%v", netbox.GetBool(obj, "is_pool")))
	bb.Add("Description", netbox.GetString(obj, "description"))

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}
	doc.Status = Ptr(StatusValue(obj))

	var props []components.CustomProperty
	props = append(props, CustomProp("nbPrefix", netbox.GetString(obj, "display")))
	if v := StatusValue(obj); v != "" {
		props = append(props, CustomProp("nbStatus", v))
	}
	if v := netbox.GetNestedString(obj, "role", "display"); v != "" {
		props = append(props, CustomProp("nbRole", v))
	}
	if v := netbox.GetNestedString(obj, "vrf", "display"); v != "" {
		props = append(props, CustomProp("nbVrf", v))
	}
	if v := netbox.GetNestedString(obj, "tenant", "display"); v != "" {
		props = append(props, CustomProp("nbTenant", v))
	}
	if v := netbox.GetNestedString(obj, "vlan", "display"); v != "" {
		props = append(props, CustomProp("nbVlan", v))
	}
	if netbox.GetBool(obj, "is_pool") {
		props = append(props, CustomProp("nbIsPool", "true"))
	}
	doc.CustomProperties = props

	return doc
}

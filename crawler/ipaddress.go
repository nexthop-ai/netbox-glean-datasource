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

type IPAddressCrawler struct{}

func init() { Register(&IPAddressCrawler{}) }

func (c *IPAddressCrawler) ObjectType() string   { return "IPAddress" }
func (c *IPAddressCrawler) DisplayLabel() string { return "IP Address" }
func (c *IPAddressCrawler) Endpoint() string     { return "/api/ipam/ip-addresses/" }

func (c *IPAddressCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("IPAddress"),
		DisplayLabel: Ptr("IP Address"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbStatus", "Status", components.PropertyTypePicklist, 1),
			FacetDef("nbRole", "Role", components.PropertyTypePicklist, 2),
			FacetDef("nbVrf", "VRF", components.PropertyTypePicklist, 3),
			FacetDef("nbTenant", "Tenant", components.PropertyTypePicklist, 4),
			PropertyDef("nbDnsName", "DNS Name", components.PropertyTypeText),
			PropertyDef("nbAssignedTo", "Assigned To", components.PropertyTypeText),
			PropertyDef("nbAddress", "Address", components.PropertyTypeText),
		},
	}
}

func (c *IPAddressCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("IPAddress", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Address", netbox.GetString(obj, "display"))
	bb.Add("Status", StatusValue(obj))
	bb.Add("DNS Name", netbox.GetString(obj, "dns_name"))
	bb.AddNested("VRF", obj, "vrf", "display")
	bb.AddNested("Tenant", obj, "tenant", "display")
	bb.AddNested("Assigned To", obj, "assigned_object", "display")
	bb.Add("Description", netbox.GetString(obj, "description"))

	// Use role if it's an object with a label.
	roleVal := ""
	if r := netbox.GetNested(obj, "role"); r != nil {
		roleVal = netbox.GetString(r, "label")
	}
	bb.Add("Role", roleVal)

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}
	doc.Status = Ptr(StatusValue(obj))

	var props []components.CustomProperty
	props = append(props, CustomProp("nbAddress", netbox.GetString(obj, "display")))
	if v := StatusValue(obj); v != "" {
		props = append(props, CustomProp("nbStatus", v))
	}
	if roleVal != "" {
		props = append(props, CustomProp("nbRole", roleVal))
	}
	if v := netbox.GetNestedString(obj, "vrf", "display"); v != "" {
		props = append(props, CustomProp("nbVrf", v))
	}
	if v := netbox.GetNestedString(obj, "tenant", "display"); v != "" {
		props = append(props, CustomProp("nbTenant", v))
	}
	if v := netbox.GetString(obj, "dns_name"); v != "" {
		props = append(props, CustomProp("nbDnsName", v))
	}
	if v := netbox.GetNestedString(obj, "assigned_object", "display"); v != "" {
		props = append(props, CustomProp("nbAssignedTo", v))
	}
	doc.CustomProperties = props

	return doc
}

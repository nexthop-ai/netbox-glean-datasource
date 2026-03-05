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

type VRFCrawler struct{}

func init() { Register(&VRFCrawler{}) }

func (c *VRFCrawler) ObjectType() string   { return "VRF" }
func (c *VRFCrawler) DisplayLabel() string { return "VRF" }
func (c *VRFCrawler) Endpoint() string     { return "/api/ipam/vrfs/" }

func (c *VRFCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("VRF"),
		DisplayLabel: Ptr("VRF"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbTenant", "Tenant", components.PropertyTypePicklist, 1),
			PropertyDef("nbRd", "Route Distinguisher", components.PropertyTypeText),
			PropertyDef("nbEnforceUnique", "Enforce Unique", components.PropertyTypeText),
		},
	}
}

func (c *VRFCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("VRF", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.Add("RD", netbox.GetString(obj, "rd"))
	bb.AddNested("Tenant", obj, "tenant", "display")
	bb.Add("Enforce Unique", fmt.Sprintf("%v", netbox.GetBool(obj, "enforce_unique")))
	bb.Add("Description", netbox.GetString(obj, "description"))

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}

	var props []components.CustomProperty
	if v := netbox.GetNestedString(obj, "tenant", "display"); v != "" {
		props = append(props, CustomProp("nbTenant", v))
	}
	if v := netbox.GetString(obj, "rd"); v != "" {
		props = append(props, CustomProp("nbRd", v))
	}
	if netbox.GetBool(obj, "enforce_unique") {
		props = append(props, CustomProp("nbEnforceUnique", "true"))
	}
	doc.CustomProperties = props

	return doc
}

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

type CircuitCrawler struct{}

func init() { Register(&CircuitCrawler{}) }

func (c *CircuitCrawler) ObjectType() string   { return "Circuit" }
func (c *CircuitCrawler) DisplayLabel() string { return "Circuit" }
func (c *CircuitCrawler) Endpoint() string     { return "/api/circuits/circuits/" }

func (c *CircuitCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("Circuit"),
		DisplayLabel: Ptr("Circuit"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbStatus", "Status", components.PropertyTypePicklist, 1),
			FacetDef("nbCircuitType", "Type", components.PropertyTypePicklist, 2),
			FacetDef("nbProvider", "Provider", components.PropertyTypePicklist, 3),
			FacetDef("nbTenant", "Tenant", components.PropertyTypePicklist, 4),
			PropertyDef("nbCid", "Circuit ID", components.PropertyTypeText),
			PropertyDef("nbCommitRate", "Commit Rate (Kbps)", components.PropertyTypeText),
		},
	}
}

func (c *CircuitCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("Circuit", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Circuit ID", netbox.GetString(obj, "cid"))
	bb.Add("Status", StatusValue(obj))
	bb.AddNested("Provider", obj, "provider", "display")
	bb.AddNested("Type", obj, "type", "display")
	bb.AddNested("Tenant", obj, "tenant", "display")
	if v := netbox.GetInt(obj, "commit_rate"); v > 0 {
		bb.Add("Commit Rate", fmt.Sprintf("%d Kbps", v))
	}
	bb.Add("Install Date", netbox.GetString(obj, "install_date"))
	bb.Add("Termination Date", netbox.GetString(obj, "termination_date"))
	bb.Add("Description", netbox.GetString(obj, "description"))

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}
	doc.Status = Ptr(StatusValue(obj))

	var props []components.CustomProperty
	if v := netbox.GetString(obj, "cid"); v != "" {
		props = append(props, CustomProp("nbCid", v))
	}
	if v := StatusValue(obj); v != "" {
		props = append(props, CustomProp("nbStatus", v))
	}
	if v := netbox.GetNestedString(obj, "type", "display"); v != "" {
		props = append(props, CustomProp("nbCircuitType", v))
	}
	if v := netbox.GetNestedString(obj, "provider", "display"); v != "" {
		props = append(props, CustomProp("nbProvider", v))
	}
	if v := netbox.GetNestedString(obj, "tenant", "display"); v != "" {
		props = append(props, CustomProp("nbTenant", v))
	}
	if v := netbox.GetInt(obj, "commit_rate"); v > 0 {
		props = append(props, CustomProp("nbCommitRate", fmt.Sprintf("%d", v)))
	}
	doc.CustomProperties = props

	return doc
}

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

type ClusterCrawler struct{}

func init() { Register(&ClusterCrawler{}) }

func (c *ClusterCrawler) ObjectType() string   { return "Cluster" }
func (c *ClusterCrawler) DisplayLabel() string { return "Cluster" }
func (c *ClusterCrawler) Endpoint() string     { return "/api/virtualization/clusters/" }

func (c *ClusterCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("Cluster"),
		DisplayLabel: Ptr("Cluster"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbStatus", "Status", components.PropertyTypePicklist, 1),
			FacetDef("nbClusterType", "Type", components.PropertyTypePicklist, 2),
			FacetDef("nbGroup", "Group", components.PropertyTypePicklist, 3),
			FacetDef("nbTenant", "Tenant", components.PropertyTypePicklist, 4),
			PropertyDef("nbSite", "Site", components.PropertyTypeText),
		},
	}
}

func (c *ClusterCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("Cluster", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.Add("Status", StatusValue(obj))
	bb.AddNested("Type", obj, "type", "display")
	bb.AddNested("Group", obj, "group", "display")
	bb.AddNested("Site", obj, "site", "display")
	bb.AddNested("Tenant", obj, "tenant", "display")
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
	if v := netbox.GetNestedString(obj, "type", "display"); v != "" {
		props = append(props, CustomProp("nbClusterType", v))
	}
	if v := netbox.GetNestedString(obj, "group", "display"); v != "" {
		props = append(props, CustomProp("nbGroup", v))
	}
	if v := netbox.GetNestedString(obj, "tenant", "display"); v != "" {
		props = append(props, CustomProp("nbTenant", v))
	}
	if v := netbox.GetNestedString(obj, "site", "display"); v != "" {
		props = append(props, CustomProp("nbSite", v))
	}
	doc.CustomProperties = props

	return doc
}

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

type VMCrawler struct{}

func init() { Register(&VMCrawler{}) }

func (c *VMCrawler) ObjectType() string   { return "VirtualMachine" }
func (c *VMCrawler) DisplayLabel() string { return "Virtual Machine" }
func (c *VMCrawler) Endpoint() string     { return "/api/virtualization/virtual-machines/" }

func (c *VMCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("VirtualMachine"),
		DisplayLabel: Ptr("Virtual Machine"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbStatus", "Status", components.PropertyTypePicklist, 1),
			FacetDef("nbCluster", "Cluster", components.PropertyTypePicklist, 2),
			FacetDef("nbSite", "Site", components.PropertyTypePicklist, 3),
			FacetDef("nbRole", "Role", components.PropertyTypePicklist, 4),
			PropertyDef("nbTenant", "Tenant", components.PropertyTypeText),
			PropertyDef("nbPlatform", "Platform", components.PropertyTypeText),
			PropertyDef("nbVcpus", "vCPUs", components.PropertyTypeText),
			PropertyDef("nbMemory", "Memory (MB)", components.PropertyTypeText),
			PropertyDef("nbDisk", "Disk (GB)", components.PropertyTypeText),
			PropertyDef("nbPrimaryIP", "Primary IP", components.PropertyTypeText),
		},
	}
}

func (c *VMCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("VirtualMachine", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.Add("Status", StatusValue(obj))
	bb.AddNested("Cluster", obj, "cluster", "display")
	bb.AddNested("Site", obj, "site", "display")
	bb.AddNested("Role", obj, "role", "display")
	bb.AddNested("Tenant", obj, "tenant", "display")
	bb.AddNested("Platform", obj, "platform", "display")
	if v := netbox.GetFloat64(obj, "vcpus"); v > 0 {
		bb.Add("vCPUs", fmt.Sprintf("%.1f", v))
	}
	if v := netbox.GetInt(obj, "memory"); v > 0 {
		bb.Add("Memory", fmt.Sprintf("%d MB", v))
	}
	if v := netbox.GetInt(obj, "disk"); v > 0 {
		bb.Add("Disk", fmt.Sprintf("%d GB", v))
	}
	bb.AddNested("Primary IPv4", obj, "primary_ip4", "display")
	bb.AddNested("Primary IPv6", obj, "primary_ip6", "display")
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
	if v := netbox.GetNestedString(obj, "cluster", "display"); v != "" {
		props = append(props, CustomProp("nbCluster", v))
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
	if v := netbox.GetNestedString(obj, "platform", "display"); v != "" {
		props = append(props, CustomProp("nbPlatform", v))
	}
	if v := netbox.GetFloat64(obj, "vcpus"); v > 0 {
		props = append(props, CustomProp("nbVcpus", fmt.Sprintf("%.1f", v)))
	}
	if v := netbox.GetInt(obj, "memory"); v > 0 {
		props = append(props, CustomProp("nbMemory", fmt.Sprintf("%d", v)))
	}
	if v := netbox.GetInt(obj, "disk"); v > 0 {
		props = append(props, CustomProp("nbDisk", fmt.Sprintf("%d", v)))
	}
	if v := netbox.GetNestedString(obj, "primary_ip4", "display"); v != "" {
		props = append(props, CustomProp("nbPrimaryIP", v))
	} else if v := netbox.GetNestedString(obj, "primary_ip6", "display"); v != "" {
		props = append(props, CustomProp("nbPrimaryIP", v))
	}
	doc.CustomProperties = props

	return doc
}

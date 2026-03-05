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

type DeviceCrawler struct{}

func init() { Register(&DeviceCrawler{}) }

func (d *DeviceCrawler) ObjectType() string   { return "Device" }
func (d *DeviceCrawler) DisplayLabel() string { return "Device" }
func (d *DeviceCrawler) Endpoint() string     { return "/api/dcim/devices/" }

func (d *DeviceCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("Device"),
		DisplayLabel: Ptr("Device"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbStatus", "Status", components.PropertyTypePicklist, 1),
			FacetDef("nbSite", "Site", components.PropertyTypePicklist, 2),
			FacetDef("nbRole", "Role", components.PropertyTypePicklist, 3),
			FacetDef("nbTenant", "Tenant", components.PropertyTypePicklist, 4),
			PropertyDef("nbDeviceType", "Device Type", components.PropertyTypeText),
			PropertyDef("nbPlatform", "Platform", components.PropertyTypeText),
			PropertyDef("nbRack", "Rack", components.PropertyTypeText),
			PropertyDef("nbSerialNumber", "Serial Number", components.PropertyTypeText),
			PropertyDef("nbAssetTag", "Asset Tag", components.PropertyTypeText),
			PropertyDef("nbPrimaryIP", "Primary IP", components.PropertyTypeText),
		},
	}
}

func (d *DeviceCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("Device", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.Add("Status", StatusValue(obj))
	bb.AddNested("Site", obj, "site", "display")
	bb.AddNested("Location", obj, "location", "display")
	bb.AddNested("Rack", obj, "rack", "display")
	bb.Add("Position", fmt.Sprintf("%v", netbox.GetFloat64(obj, "position")))
	bb.AddNested("Role", obj, "role", "display")
	bb.AddNested("Device Type", obj, "device_type", "display")
	bb.AddNested("Manufacturer", obj, "device_type", "manufacturer", "display")
	bb.AddNested("Platform", obj, "platform", "display")
	bb.AddNested("Tenant", obj, "tenant", "display")
	bb.Add("Serial Number", netbox.GetString(obj, "serial"))
	bb.Add("Asset Tag", netbox.GetString(obj, "asset_tag"))
	bb.AddNested("Primary IPv4", obj, "primary_ip4", "display")
	bb.AddNested("Primary IPv6", obj, "primary_ip6", "display")
	bb.AddNested("OOB IP", obj, "oob_ip", "display")
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
	if v := netbox.GetNestedString(obj, "device_type", "display"); v != "" {
		props = append(props, CustomProp("nbDeviceType", v))
	}
	if v := netbox.GetNestedString(obj, "platform", "display"); v != "" {
		props = append(props, CustomProp("nbPlatform", v))
	}
	if v := netbox.GetNestedString(obj, "rack", "display"); v != "" {
		props = append(props, CustomProp("nbRack", v))
	}
	if v := netbox.GetString(obj, "serial"); v != "" {
		props = append(props, CustomProp("nbSerialNumber", v))
	}
	if v := netbox.GetString(obj, "asset_tag"); v != "" {
		props = append(props, CustomProp("nbAssetTag", v))
	}
	if v := netbox.GetNestedString(obj, "primary_ip4", "display"); v != "" {
		props = append(props, CustomProp("nbPrimaryIP", v))
	} else if v := netbox.GetNestedString(obj, "primary_ip6", "display"); v != "" {
		props = append(props, CustomProp("nbPrimaryIP", v))
	}
	doc.CustomProperties = props

	return doc
}

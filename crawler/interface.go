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
	"strings"

	"github.com/gleanwork/api-client-go/models/components"
	"github.com/nexthop-ai/netbox-glean-datasource/netbox"
)

type InterfaceCrawler struct{}

func init() { Register(&InterfaceCrawler{}) }

func (c *InterfaceCrawler) ObjectType() string   { return "Interface" }
func (c *InterfaceCrawler) DisplayLabel() string { return "Interface" }
func (c *InterfaceCrawler) Endpoint() string     { return "/api/dcim/interfaces/" }

func (c *InterfaceCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("Interface"),
		DisplayLabel: Ptr("Interface"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbDevice", "Device", components.PropertyTypePicklist, 1),
			FacetDef("nbInterfaceType", "Type", components.PropertyTypePicklist, 2),
			PropertyDef("nbSpeed", "Speed", components.PropertyTypeText),
			PropertyDef("nbMtu", "MTU", components.PropertyTypeText),
			PropertyDef("nbMode", "Mode", components.PropertyTypeText),
			PropertyDef("nbMacAddress", "MAC Address", components.PropertyTypeText),
			PropertyDef("nbEnabled", "Enabled", components.PropertyTypeText),
			PropertyDef("nbConnectedTo", "Connected To", components.PropertyTypeText),
		},
	}
}

func (c *InterfaceCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("Interface", obj, datasource, netboxURL)

	ifType := netbox.GetNested(obj, "type")
	typeLabel := ""
	if ifType != nil {
		typeLabel = netbox.GetString(ifType, "label")
	}

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.AddNested("Device", obj, "device", "display")
	bb.Add("Type", typeLabel)
	bb.Add("Status", StatusValue(obj))
	bb.Add("Enabled", fmt.Sprintf("%v", netbox.GetBool(obj, "enabled")))
	if v := netbox.GetInt(obj, "speed"); v > 0 {
		bb.Add("Speed", fmt.Sprintf("%d Kbps", v))
	}
	if v := netbox.GetInt(obj, "mtu"); v > 0 {
		bb.Add("MTU", fmt.Sprintf("%d", v))
	}
	bb.Add("MAC Address", netbox.GetString(obj, "mac_address"))
	if mode := netbox.GetNested(obj, "mode"); mode != nil {
		bb.Add("Mode", netbox.GetString(mode, "label"))
	}
	bb.Add("Description", netbox.GetString(obj, "description"))

	// Connection/cable info.
	connectedTo := connectedEndpoints(obj)
	if connectedTo != "" {
		bb.Add("Connected To", connectedTo)
	}
	if cable := netbox.GetNested(obj, "cable"); cable != nil {
		bb.Add("Cable", netbox.GetString(cable, "label"))
	}

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}
	doc.Status = Ptr(StatusValue(obj))

	// Set container to the parent device.
	if device := netbox.GetNestedString(obj, "device", "display"); device != "" {
		doc.Container = Ptr(device)
	}

	var props []components.CustomProperty
	if v := netbox.GetNestedString(obj, "device", "display"); v != "" {
		props = append(props, CustomProp("nbDevice", v))
	}
	if typeLabel != "" {
		props = append(props, CustomProp("nbInterfaceType", typeLabel))
	}
	if v := StatusValue(obj); v != "" {
		props = append(props, CustomProp("nbStatus", v))
	}
	if v := netbox.GetInt(obj, "speed"); v > 0 {
		props = append(props, CustomProp("nbSpeed", fmt.Sprintf("%d", v)))
	}
	if v := netbox.GetInt(obj, "mtu"); v > 0 {
		props = append(props, CustomProp("nbMtu", fmt.Sprintf("%d", v)))
	}
	if v := netbox.GetString(obj, "mac_address"); v != "" {
		props = append(props, CustomProp("nbMacAddress", v))
	}
	if mode := netbox.GetNested(obj, "mode"); mode != nil {
		if v := netbox.GetString(mode, "label"); v != "" {
			props = append(props, CustomProp("nbMode", v))
		}
	}
	props = append(props, CustomProp("nbEnabled", fmt.Sprintf("%v", netbox.GetBool(obj, "enabled"))))
	if connectedTo != "" {
		props = append(props, CustomProp("nbConnectedTo", connectedTo))
	}
	doc.CustomProperties = props

	return doc
}

// connectedEndpoints extracts "device:port" strings from the connected_endpoints array.
func connectedEndpoints(obj map[string]any) string {
	eps, ok := obj["connected_endpoints"]
	if !ok || eps == nil {
		return ""
	}
	arr, ok := eps.([]any)
	if !ok || len(arr) == 0 {
		return ""
	}
	var parts []string
	for _, ep := range arr {
		m, ok := ep.(map[string]any)
		if !ok {
			continue
		}
		device := netbox.GetNestedString(m, "device", "display")
		port := netbox.GetString(m, "display")
		if device != "" && port != "" {
			parts = append(parts, device+":"+port)
		} else if port != "" {
			parts = append(parts, port)
		}
	}
	return strings.Join(parts, ", ")
}

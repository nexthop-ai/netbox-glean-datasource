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

// PowerPortCrawler indexes device power ports (e.g., PSU inlets). The
// connected_endpoints field captures which PDU outlet each PSU is plugged
// into, so this provides the PSU -> PDU/outlet mapping.
type PowerPortCrawler struct{}

func init() { Register(&PowerPortCrawler{}) }

func (c *PowerPortCrawler) ObjectType() string   { return "PowerPort" }
func (c *PowerPortCrawler) DisplayLabel() string { return "Power Port" }
func (c *PowerPortCrawler) Endpoint() string     { return "/api/dcim/power-ports/" }

func (c *PowerPortCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("PowerPort"),
		DisplayLabel: Ptr("Power Port"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbDevice", "Device", components.PropertyDefinitionPropertyTypePicklist, 1),
			FacetDef("nbPortType", "Type", components.PropertyDefinitionPropertyTypePicklist, 2),
			PropertyDef("nbLabel", "Label", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbMaximumDraw", "Maximum Draw (W)", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbAllocatedDraw", "Allocated Draw (W)", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbConnectedTo", "Connected To", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbCable", "Cable", components.PropertyDefinitionPropertyTypeText),
		},
	}
}

func (c *PowerPortCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("PowerPort", obj, datasource, netboxURL)

	typeLabel := ""
	if portType := netbox.GetNested(obj, "type"); portType != nil {
		typeLabel = netbox.GetString(portType, "label")
	}

	maximumDraw := ""
	if v, ok := obj["maximum_draw"].(float64); ok {
		maximumDraw = fmt.Sprintf("%d", int64(v))
	}
	allocatedDraw := ""
	if v, ok := obj["allocated_draw"].(float64); ok {
		allocatedDraw = fmt.Sprintf("%d", int64(v))
	}

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.AddNested("Device", obj, "device", "display")
	bb.Add("Type", typeLabel)
	bb.Add("Label", netbox.GetString(obj, "label"))
	bb.Add("Maximum Draw (W)", maximumDraw)
	bb.Add("Allocated Draw (W)", allocatedDraw)
	bb.Add("Description", netbox.GetString(obj, "description"))

	// Connection/cable info: which PDU outlet this PSU is plugged into.
	connectedTo := connectedEndpoints(obj)
	bb.Add("Connected To", connectedTo)
	cableLabel := ""
	if cable := netbox.GetNested(obj, "cable"); cable != nil {
		cableLabel = netbox.GetString(cable, "label")
		if cableLabel == "" {
			cableLabel = netbox.GetString(cable, "display")
		}
	}
	bb.Add("Cable", cableLabel)

	if netbox.GetBool(obj, "mark_connected") {
		bb.Add("Mark Connected", "Yes")
	}

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}

	// Set container to the parent device
	if device := netbox.GetNestedString(obj, "device", "display"); device != "" {
		doc.Container = Ptr(device)
	}

	var props []components.CustomProperty
	if v := netbox.GetNestedString(obj, "device", "display"); v != "" {
		props = append(props, CustomProp("nbDevice", v))
	}
	if typeLabel != "" {
		props = append(props, CustomProp("nbPortType", typeLabel))
	}
	if v := netbox.GetString(obj, "label"); v != "" {
		props = append(props, CustomProp("nbLabel", v))
	}
	if maximumDraw != "" {
		props = append(props, CustomProp("nbMaximumDraw", maximumDraw))
	}
	if allocatedDraw != "" {
		props = append(props, CustomProp("nbAllocatedDraw", allocatedDraw))
	}
	if connectedTo != "" {
		props = append(props, CustomProp("nbConnectedTo", connectedTo))
	}
	if cableLabel != "" {
		props = append(props, CustomProp("nbCable", cableLabel))
	}
	doc.CustomProperties = props

	return doc
}

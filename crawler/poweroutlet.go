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

// PowerOutletCrawler indexes PDU power outlets. The connected_endpoints
// field captures which device power port (PSU) is plugged into each
// outlet, giving the PDU outlet -> PSU mapping.
type PowerOutletCrawler struct{}

func init() { Register(&PowerOutletCrawler{}) }

func (c *PowerOutletCrawler) ObjectType() string   { return "PowerOutlet" }
func (c *PowerOutletCrawler) DisplayLabel() string { return "Power Outlet" }
func (c *PowerOutletCrawler) Endpoint() string     { return "/api/dcim/power-outlets/" }

func (c *PowerOutletCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("PowerOutlet"),
		DisplayLabel: Ptr("Power Outlet"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbDevice", "Device", components.PropertyDefinitionPropertyTypePicklist, 1),
			FacetDef("nbPortType", "Type", components.PropertyDefinitionPropertyTypePicklist, 2),
			PropertyDef("nbLabel", "Label", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbPowerPort", "Power Port", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbFeedLeg", "Feed Leg", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbConnectedTo", "Connected To", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbCable", "Cable", components.PropertyDefinitionPropertyTypeText),
		},
	}
}

func (c *PowerOutletCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("PowerOutlet", obj, datasource, netboxURL)

	typeLabel := ""
	if outletType := netbox.GetNested(obj, "type"); outletType != nil {
		typeLabel = netbox.GetString(outletType, "label")
	}

	feedLeg := ""
	if leg := netbox.GetNested(obj, "feed_leg"); leg != nil {
		feedLeg = netbox.GetString(leg, "label")
	}

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.AddNested("Device", obj, "device", "display")
	bb.Add("Type", typeLabel)
	bb.Add("Label", netbox.GetString(obj, "label"))
	bb.AddNested("Power Port", obj, "power_port", "display")
	bb.Add("Feed Leg", feedLeg)
	bb.Add("Description", netbox.GetString(obj, "description"))

	// Connection/cable info: which device PSU is plugged into this outlet.
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

	// Set container to the parent device (the PDU)
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
	if v := netbox.GetNestedString(obj, "power_port", "display"); v != "" {
		props = append(props, CustomProp("nbPowerPort", v))
	}
	if feedLeg != "" {
		props = append(props, CustomProp("nbFeedLeg", feedLeg))
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

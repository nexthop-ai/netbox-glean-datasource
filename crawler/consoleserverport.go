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

type ConsoleServerPortCrawler struct{}

func init() { Register(&ConsoleServerPortCrawler{}) }

func (c *ConsoleServerPortCrawler) ObjectType() string   { return "ConsoleServerPort" }
func (c *ConsoleServerPortCrawler) DisplayLabel() string { return "Console Server Port" }
func (c *ConsoleServerPortCrawler) Endpoint() string     { return "/api/dcim/console-server-ports/" }

func (c *ConsoleServerPortCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("ConsoleServerPort"),
		DisplayLabel: Ptr("Console Server Port"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
		PropertyDefinitions: []components.PropertyDefinition{
			FacetDef("nbDevice", "Device", components.PropertyDefinitionPropertyTypePicklist, 1),
			FacetDef("nbPortType", "Type", components.PropertyDefinitionPropertyTypePicklist, 2),
			PropertyDef("nbSpeed", "Speed", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbLabel", "Label", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbConnectedTo", "Connected To", components.PropertyDefinitionPropertyTypeText),
			PropertyDef("nbCable", "Cable", components.PropertyDefinitionPropertyTypeText),
		},
	}
}

func (c *ConsoleServerPortCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("ConsoleServerPort", obj, datasource, netboxURL)

	portType := netbox.GetNested(obj, "type")
	typeLabel := ""
	if portType != nil {
		typeLabel = netbox.GetString(portType, "label")
	}

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.AddNested("Device", obj, "device", "display")
	bb.Add("Type", typeLabel)
	bb.Add("Label", netbox.GetString(obj, "label"))

	// Speed information
	if speed := netbox.GetNested(obj, "speed"); speed != nil {
		speedLabel := netbox.GetString(speed, "label")
		if speedLabel != "" {
			bb.Add("Speed", speedLabel)
		}
	}

	bb.Add("Description", netbox.GetString(obj, "description"))

	// Connection/cable info
	connectedTo := connectedEndpoints(obj)
	if connectedTo != "" {
		bb.Add("Connected To", connectedTo)
	}
	if cable := netbox.GetNested(obj, "cable"); cable != nil {
		cableLabel := netbox.GetString(cable, "label")
		if cableLabel == "" {
			cableLabel = netbox.GetString(cable, "display")
		}
		if cableLabel != "" {
			bb.Add("Cable", cableLabel)
		}
	}

	// Mark connected status
	markConnected := netbox.GetBool(obj, "mark_connected")
	if markConnected {
		bb.Add("Mark Connected", "Yes")
	}

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}

	// Set container to the parent device (console server)
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
	if speed := netbox.GetNested(obj, "speed"); speed != nil {
		if speedLabel := netbox.GetString(speed, "label"); speedLabel != "" {
			props = append(props, CustomProp("nbSpeed", speedLabel))
		}
	}
	if v := netbox.GetString(obj, "label"); v != "" {
		props = append(props, CustomProp("nbLabel", v))
	}
	if connectedTo != "" {
		props = append(props, CustomProp("nbConnectedTo", connectedTo))
	}
	if cable := netbox.GetNested(obj, "cable"); cable != nil {
		if cableLabel := netbox.GetString(cable, "label"); cableLabel != "" {
			props = append(props, CustomProp("nbCable", cableLabel))
		} else if cableDisplay := netbox.GetString(cable, "display"); cableDisplay != "" {
			props = append(props, CustomProp("nbCable", cableDisplay))
		}
	}
	doc.CustomProperties = props

	return doc
}

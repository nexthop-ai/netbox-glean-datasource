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

type ManufacturerCrawler struct{}

func init() { Register(&ManufacturerCrawler{}) }

func (c *ManufacturerCrawler) ObjectType() string   { return "Manufacturer" }
func (c *ManufacturerCrawler) DisplayLabel() string { return "Manufacturer" }
func (c *ManufacturerCrawler) Endpoint() string     { return "/api/dcim/manufacturers/" }

func (c *ManufacturerCrawler) ObjectDefinition() components.ObjectDefinition {
	return components.ObjectDefinition{
		Name:         Ptr("Manufacturer"),
		DisplayLabel: Ptr("Manufacturer"),
		DocCategory:  components.DocCategoryKnowledgeHub.ToPointer(),
	}
}

func (c *ManufacturerCrawler) Transform(obj map[string]any, datasource, netboxURL string) components.DocumentDefinition {
	doc := BaseDocument("Manufacturer", obj, datasource, netboxURL)

	var bb BodyBuilder
	bb.Add("Name", netbox.GetString(obj, "display"))
	bb.Add("Description", netbox.GetString(obj, "description"))

	doc.Body = &components.ContentDefinition{
		MimeType:    "text/plain",
		TextContent: Ptr(bb.String()),
	}

	return doc
}

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

package glean

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"

	apiclientgo "github.com/gleanwork/api-client-go"
	"github.com/gleanwork/api-client-go/models/components"
	"github.com/nexthop-ai/netbox-glean-datasource/crawler"
)

// RegisterDatasource creates or updates the custom datasource schema in Glean.
func RegisterDatasource(ctx context.Context, sdk *apiclientgo.Glean, dsName, displayName, netboxURL string, crawlers map[string]crawler.Crawler) error {
	objDefs := make([]components.ObjectDefinition, 0, len(crawlers))

	// Sort for deterministic ordering.
	names := make([]string, 0, len(crawlers))
	for name := range crawlers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		objDefs = append(objDefs, crawlers[name].ObjectDefinition())
	}

	connType := components.CustomDatasourceConfigConnectorTypePushAPI
	category := components.DatasourceCategoryKnowledgeHub
	urlRegex := regexp.QuoteMeta(strings.TrimRight(netboxURL, "/")) + "/.*"

	dsConfig := components.CustomDatasourceConfig{
		Name:                dsName,
		DisplayName:         &displayName,
		DatasourceCategory:  &category,
		ConnectorType:       &connType,
		HomeURL:             &netboxURL,
		URLRegex:            &urlRegex,
		ObjectDefinitions:   objDefs,
		IconURL:             crawler.Ptr("https://raw.githubusercontent.com/netbox-community/netbox/main/docs/netbox_logo_light.svg"),
		IconDarkURL:         crawler.Ptr("https://raw.githubusercontent.com/netbox-community/netbox/main/docs/netbox_logo_dark.svg"),
		IsOnPrem:            crawler.Ptr(true),
		IncludeUtmSource:    crawler.Ptr(true),
		SuggestionText:      crawler.Ptr("Search for network devices, IPs, sites..."),
	}

	slog.Info("registering datasource", "name", dsName, "objectTypes", len(objDefs))
	_, err := sdk.Indexing.Datasources.Add(ctx, dsConfig)
	if err != nil {
		return fmt.Errorf("registering datasource %q: %w", dsName, err)
	}
	slog.Info("datasource registered successfully", "name", dsName)
	return nil
}

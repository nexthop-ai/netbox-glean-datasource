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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/nexthop-ai/netbox-glean-datasource/crawler"
	"github.com/nexthop-ai/netbox-glean-datasource/netbox"

	// Register all crawlers.
	_ "github.com/nexthop-ai/netbox-glean-datasource/crawler"
)

func main() {
	url := flag.String("url", "https://netbox.internal.nexthop.ai", "NetBox URL")
	token := flag.String("token", "", "NetBox API token (or NETBOX_TOKEN env var)")
	objectType := flag.String("type", "", "only dump this object type (e.g. Device)")
	datasource := flag.String("datasource", "netbox", "datasource name for document IDs")
	limit := flag.Int("limit", 0, "max objects per type (0 = all)")
	flag.Parse()

	if *token == "" {
		*token = os.Getenv("NETBOX_TOKEN")
	}
	if *token == "" {
		fmt.Fprintln(os.Stderr, "error: --token or NETBOX_TOKEN required")
		os.Exit(1)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))

	client := netbox.NewClient(*url, *token)
	ctx := context.Background()

	crawlers := crawler.All()
	names := make([]string, 0, len(crawlers))
	for name := range crawlers {
		if *objectType != "" && name != *objectType {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	if len(names) == 0 {
		fmt.Fprintf(os.Stderr, "error: no crawler found for type %q\n", *objectType)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	for _, name := range names {
		c := crawlers[name]
		fmt.Fprintf(os.Stderr, "--- %s (%s) ---\n", c.DisplayLabel(), c.Endpoint())

		count := 0
		err := client.List(ctx, c.Endpoint(), nil, func(results []map[string]any) error {
			for _, obj := range results {
				if *limit > 0 && count >= *limit {
					return nil
				}
				doc := c.Transform(obj, *datasource, *url)
				if err := enc.Encode(doc); err != nil {
					return err
				}
				count++
			}
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching %s: %v\n", name, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "  %d documents\n", count)
	}
}

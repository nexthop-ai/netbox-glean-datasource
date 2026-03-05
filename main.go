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
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiclientgo "github.com/gleanwork/api-client-go"
	"github.com/nexthop-ai/netbox-glean-datasource/crawler"
	gleanpkg "github.com/nexthop-ai/netbox-glean-datasource/glean"
	"github.com/nexthop-ai/netbox-glean-datasource/netbox"

	// Import all crawlers to trigger init() registration.
	_ "github.com/nexthop-ai/netbox-glean-datasource/crawler"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "setup":
		cmdSetup(os.Args[2:])
	case "sync":
		cmdSync(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n\nCommands:\n  setup   Register/update the datasource schema in Glean\n  sync    Run a one-shot sync\n  serve   Run as a daemon, syncing on an interval\n", os.Args[0])
	os.Exit(1)
}

func parseFlags(args []string) (*Config, string) {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "path to config file")
	logLevel := fs.String("log-level", "info", "log level (debug, info, warn, error)")
	fs.Parse(args)

	setupLogging(*logLevel)

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}
	return cfg, fs.Arg(0)
}

func setupLogging(level string) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})))
}

func newGleanSDK(cfg *Config) *apiclientgo.Glean {
	return apiclientgo.New(
		apiclientgo.WithInstance(cfg.Glean.Instance),
		apiclientgo.WithSecurity(cfg.Glean.Token),
	)
}

func newSyncer(cfg *Config, sdk *apiclientgo.Glean) *gleanpkg.Syncer {
	return &gleanpkg.Syncer{
		GleanSDK:    sdk,
		NetBox:      netbox.NewClient(cfg.NetBox.URL, cfg.NetBox.Token),
		Datasource:  cfg.Datasource.Name,
		NetBoxURL:   cfg.NetBox.URL,
		BatchSize:   cfg.Sync.BatchSize,
		Concurrency: cfg.Sync.Concurrency,
	}
}

func cmdSetup(args []string) {
	cfg, _ := parseFlags(args)
	sdk := newGleanSDK(cfg)

	ctx := context.Background()
	if err := gleanpkg.RegisterDatasource(ctx, sdk, cfg.Datasource.Name, cfg.Datasource.DisplayName, cfg.NetBox.URL, crawler.All()); err != nil {
		slog.Error("setup failed", "error", err)
		os.Exit(1)
	}
}

func cmdSync(args []string) {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "path to config file")
	logLevel := fs.String("log-level", "info", "log level")
	since := fs.String("since", "", "incremental sync from this RFC3339 timestamp")
	fs.Parse(args)

	setupLogging(*logLevel)

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	sdk := newGleanSDK(cfg)
	syncer := newSyncer(cfg, sdk)

	var sinceTime *time.Time
	if *since != "" {
		t, err := time.Parse(time.RFC3339, *since)
		if err != nil {
			slog.Error("invalid --since timestamp", "error", err)
			os.Exit(1)
		}
		sinceTime = &t
	}

	ctx := context.Background()
	if err := syncer.SyncAll(ctx, cfg.Sync.ObjectTypes, sinceTime); err != nil {
		slog.Error("sync failed", "error", err)
		os.Exit(1)
	}
}

func cmdServe(args []string) {
	cfg, _ := parseFlags(args)
	sdk := newGleanSDK(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Register datasource on startup.
	if err := gleanpkg.RegisterDatasource(ctx, sdk, cfg.Datasource.Name, cfg.Datasource.DisplayName, cfg.NetBox.URL, crawler.All()); err != nil {
		slog.Error("setup failed", "error", err)
		os.Exit(1)
	}

	syncer := newSyncer(cfg, sdk)
	syncCount := 0

	for {
		var since *time.Time
		if syncCount > 0 && syncCount%cfg.Sync.FullSyncEvery != 0 {
			t := time.Now().Add(-cfg.Sync.Interval - time.Minute) // small overlap for safety
			since = &t
		}

		syncType := "full"
		if since != nil {
			syncType = "incremental"
		}
		slog.Info("starting sync cycle", "cycle", syncCount+1, "type", syncType)

		if err := syncer.SyncAll(ctx, cfg.Sync.ObjectTypes, since); err != nil {
			slog.Error("sync cycle failed", "cycle", syncCount+1, "error", err)
		} else {
			slog.Info("sync cycle completed", "cycle", syncCount+1)
		}
		syncCount++

		slog.Info("sleeping until next sync", "interval", cfg.Sync.Interval)
		select {
		case <-ctx.Done():
			slog.Info("shutting down")
			return
		case <-time.After(cfg.Sync.Interval):
		}
	}
}

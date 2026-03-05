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
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NetBox struct {
		URL   string `yaml:"url"`
		Token string `yaml:"token"`
	} `yaml:"netbox"`
	Glean struct {
		Instance string `yaml:"instance"`
		Token    string `yaml:"token"`
	} `yaml:"glean"`
	Datasource struct {
		Name        string `yaml:"name"`
		DisplayName string `yaml:"display_name"`
	} `yaml:"datasource"`
	Sync struct {
		ObjectTypes    []string      `yaml:"object_types"`
		BatchSize      int           `yaml:"batch_size"`
		Concurrency    int           `yaml:"concurrency"`
		Interval       time.Duration `yaml:"interval"`
		FullSyncEvery  int           `yaml:"full_sync_every"`
	} `yaml:"sync"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}

	// Environment variable overrides.
	if v := os.Getenv("NETBOX_URL"); v != "" {
		cfg.NetBox.URL = v
	}
	if v := os.Getenv("NETBOX_TOKEN"); v != "" {
		cfg.NetBox.Token = v
	}
	if v := os.Getenv("GLEAN_INSTANCE"); v != "" {
		cfg.Glean.Instance = v
	}
	if v := os.Getenv("GLEAN_TOKEN"); v != "" {
		cfg.Glean.Token = v
	}

	// Defaults.
	if cfg.Datasource.Name == "" {
		cfg.Datasource.Name = "netbox"
	}
	if cfg.Datasource.DisplayName == "" {
		cfg.Datasource.DisplayName = "NetBox"
	}
	if cfg.Sync.BatchSize <= 0 {
		cfg.Sync.BatchSize = 100
	}
	if cfg.Sync.Concurrency <= 0 {
		cfg.Sync.Concurrency = 4
	}
	if cfg.Sync.Interval <= 0 {
		cfg.Sync.Interval = time.Hour
	}
	if cfg.Sync.FullSyncEvery <= 0 {
		cfg.Sync.FullSyncEvery = 24
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.NetBox.URL == "" {
		return fmt.Errorf("netbox.url is required")
	}
	if c.NetBox.Token == "" {
		return fmt.Errorf("netbox.token is required")
	}
	if c.Glean.Instance == "" {
		return fmt.Errorf("glean.instance is required")
	}
	if c.Glean.Token == "" {
		return fmt.Errorf("glean.token is required")
	}
	return nil
}

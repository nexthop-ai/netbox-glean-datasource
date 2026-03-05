# NetBox Glean Datasource

A custom datasource connector for [Glean](https://glean.com) that crawls and
indexes data from [NetBox](https://netbox.dev), making network infrastructure
information searchable through Glean's unified search and AI platform.

## What it does

This connector bridges NetBox (a network infrastructure management system) and
Glean (an enterprise search platform) by:

- Crawling NetBox infrastructure data via its REST API
- Transforming it into Glean-compatible documents with searchable metadata
- Pushing the indexed content to Glean using the Indexing (Push) API
- Running as a long-lived daemon that periodically syncs changes

## Indexed object types

| Category | Object Types |
|---|---|
| DCIM | Device, Site, Rack, Interface, Manufacturer, Platform |
| IPAM | IP Address, Prefix, VLAN, VRF |
| Circuits | Circuit, Provider |
| Virtualization | Virtual Machine, Cluster |
| Tenancy | Tenant |

Each object is indexed with its key fields as searchable text and custom
properties for faceted filtering (e.g., filter by site, role, tenant, status).

## Configuration

Copy `config.yaml.example` to `config.yaml` and fill in:

```yaml
netbox:
  url: https://netbox.example.com
  token: your-netbox-api-token

glean:
  instance: your-glean-instance   # just the subdomain, e.g., "acme"
  token: your-glean-indexing-api-token

datasource:
  name: netbox
  display_name: NetBox

sync:
  batch_size: 100       # documents per Glean API call
  concurrency: 4        # parallel NetBox fetches
  interval: 1h          # sync interval in serve mode
  full_sync_every: 24   # full sync every N cycles (vs incremental)
```

All settings can be overridden with environment variables:
`NETBOX_URL`, `NETBOX_TOKEN`, `GLEAN_INSTANCE`, `GLEAN_TOKEN`.

## Usage

### Register the datasource schema in Glean

```sh
./netbox-glean-datasource setup --config config.yaml
```

### Run a one-shot sync

```sh
./netbox-glean-datasource sync --config config.yaml
```

For incremental sync (only objects updated since a given time):

```sh
./netbox-glean-datasource sync --config config.yaml --since 2024-01-01T00:00:00Z
```

### Run as a daemon

```sh
./netbox-glean-datasource serve --config config.yaml
```

This registers the datasource, performs an initial full sync, then syncs
periodically at the configured interval. Incremental syncs are used between
full syncs (controlled by `full_sync_every`). Handles SIGINT/SIGTERM for
graceful shutdown.

In serve mode, an HTTP server is started (default `:8080`) providing:

- `/` — Status page showing sync state, document counts, and uptime
- `/metrics` — Prometheus metrics endpoint
- `/sync` — Trigger an immediate sync cycle

## Docker

```sh
docker build -t netbox-glean-datasource .
docker run -v $(pwd)/config.yaml:/config.yaml netbox-glean-datasource serve --config /config.yaml
```

## Development

### Debug tool

The `cmd/netboxdump` tool dumps what would be indexed without touching Glean:

```sh
go run ./cmd/netboxdump/ --token <NETBOX_TOKEN> --url https://netbox.example.com --limit 5
```

Options: `--type Device` (single type), `--limit N` (max per type).

### Tests

```sh
go test ./...
```

Tests use `net/http/httptest` mock servers and don't require access to NetBox
or Glean.

## License

Copyright (c) 2026-present, Nexthop Systems, Inc.

Licensed under the Apache License, Version 2.0.

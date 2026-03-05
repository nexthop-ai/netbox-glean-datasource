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

import "github.com/prometheus/client_golang/prometheus"

var (
	lastSyncSuccessTimestamp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "netbox_glean_last_sync_success_timestamp",
		Help: "Unix timestamp of the last successful sync.",
	}, []string{"sync_type"})

	lastSyncSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "netbox_glean_last_sync_success",
		Help: "Whether the last sync was successful (1) or not (0).",
	})

	documentCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "netbox_glean_document_count",
		Help: "Number of documents indexed per object type.",
	}, []string{"object_type"})

	syncCyclesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "netbox_glean_sync_cycles_total",
		Help: "Total number of sync cycles attempted.",
	}, []string{"sync_type"})

	syncErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "netbox_glean_sync_errors_total",
		Help: "Total number of failed sync cycles.",
	}, []string{"sync_type", "error_source"})

	syncDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "netbox_glean_sync_duration_seconds",
		Help:    "Duration of sync cycles in seconds.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12), // 1s to ~68min
	}, []string{"sync_type"})

	netboxFetchDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "netbox_glean_netbox_fetch_duration_seconds",
		Help:    "Duration of the NetBox fetch phase in seconds.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1s to ~17min
	})

	gleanPushDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "netbox_glean_glean_push_duration_seconds",
		Help:    "Duration of the Glean push phase in seconds.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	})

	documentsPerSync = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "netbox_glean_documents_pushed_total",
		Help: "Total number of documents pushed to Glean.",
	}, []string{"object_type"})

	lastSyncUpdatedCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "netbox_glean_last_sync_updated_count",
		Help: "Number of documents updated in the last sync cycle per object type.",
	}, []string{"object_type"})
)

func init() {
	prometheus.MustRegister(
		lastSyncSuccessTimestamp,
		lastSyncSuccess,
		documentCount,
		syncCyclesTotal,
		syncErrorsTotal,
		syncDuration,
		netboxFetchDuration,
		gleanPushDuration,
		documentsPerSync,
		lastSyncUpdatedCount,
	)
}

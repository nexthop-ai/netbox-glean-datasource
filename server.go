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
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SyncStatus tracks the current state of the daemon for the status page.
type SyncStatus struct {
	mu              sync.RWMutex
	lastSyncTime    time.Time
	lastSyncType    string
	lastSyncError   error
	lastSyncDocs    map[string]int
	syncCount       int
	syncing         bool
	startTime       time.Time
	triggerSyncChan chan struct{}
}

func NewSyncStatus(triggerChan chan struct{}) *SyncStatus {
	return &SyncStatus{
		startTime:       time.Now(),
		triggerSyncChan: triggerChan,
	}
}

func (s *SyncStatus) SetSyncing(syncing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncing = syncing
}

func (s *SyncStatus) RecordSync(syncType string, docs map[string]int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSyncTime = time.Now()
	s.lastSyncType = syncType
	s.lastSyncError = err
	s.lastSyncDocs = docs
	s.syncCount++
}

func startHTTPServer(addr string, status *SyncStatus) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", status.handleStatus)
	mux.HandleFunc("/sync", status.handleTriggerSync)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()
	return srv
}

func (s *SyncStatus) handleStatus(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `<!DOCTYPE html>
<html><head><title>NetBox Glean Datasource</title>
<style>
body { font-family: sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px; }
table { border-collapse: collapse; width: 100%%; margin: 16px 0; }
th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid #ddd; }
th { background: #f5f5f5; }
.ok { color: green; } .err { color: red; } .syncing { color: orange; }
a.btn { display: inline-block; padding: 8px 16px; background: #0066cc; color: white; text-decoration: none; border-radius: 4px; margin: 8px 0; }
a.btn:hover { background: #0052a3; }
</style></head><body>
<h1>NetBox Glean Datasource</h1>
<p>Uptime: %s</p>
`, time.Since(s.startTime).Round(time.Second))

	if s.syncing {
		fmt.Fprintf(&buf, `<p class="syncing"><strong>Status: Syncing...</strong></p>`)
	} else if s.syncCount == 0 {
		fmt.Fprintf(&buf, `<p>Status: No sync completed yet</p>`)
	} else if s.lastSyncError != nil {
		fmt.Fprintf(&buf, `<p class="err"><strong>Status: Last sync failed</strong><br>Error: %s</p>`, s.lastSyncError)
	} else {
		fmt.Fprintf(&buf, `<p class="ok"><strong>Status: OK</strong></p>`)
	}

	if s.syncCount > 0 {
		fmt.Fprintf(&buf, `<p>Last sync: %s (%s, %s ago)</p>`,
			s.lastSyncType,
			s.lastSyncTime.Format(time.RFC3339),
			time.Since(s.lastSyncTime).Round(time.Second))
		fmt.Fprintf(&buf, `<p>Total sync cycles: %d</p>`, s.syncCount)
	}

	if len(s.lastSyncDocs) > 0 {
		fmt.Fprintf(&buf, `<h2>Documents (last sync)</h2><table><tr><th>Object Type</th><th>Count</th></tr>`)
		total := 0
		for objType, count := range s.lastSyncDocs {
			fmt.Fprintf(&buf, `<tr><td>%s</td><td>%d</td></tr>`, objType, count)
			total += count
		}
		fmt.Fprintf(&buf, `<tr><th>Total</th><th>%d</th></tr></table>`, total)
	} else {
		fmt.Fprintf(&buf, `<p>No document available yet, still indexing...</p>`)
	}

	fmt.Fprintf(&buf, `<h2>Actions</h2>
<a class="btn" href="/sync">Trigger Sync Now</a>
</body></html>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func (s *SyncStatus) handleTriggerSync(w http.ResponseWriter, _ *http.Request) {
	select {
	case s.triggerSyncChan <- struct{}{}:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta http-equiv="refresh" content="3;url=/"></head><body>
<p>Sync triggered. <a href="/">Redirecting to status page...</a></p></body></html>`)
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><body>
<p>A sync is already in progress or pending. <a href="/">Back to status page.</a></p></body></html>`)
	}
}

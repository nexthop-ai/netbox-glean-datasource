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
	"net/url"
	"sort"
	"time"

	apiclientgo "github.com/gleanwork/api-client-go"
	"github.com/gleanwork/api-client-go/models/components"
	"github.com/nexthop-ai/netbox-glean-datasource/crawler"
	"github.com/nexthop-ai/netbox-glean-datasource/netbox"
	"golang.org/x/sync/errgroup"
)

type Syncer struct {
	GleanSDK    *apiclientgo.Glean
	NetBox      *netbox.Client
	Datasource  string
	NetBoxURL   string
	BatchSize   int
	Concurrency int
}

// SyncResult contains per-object-type document counts and phase durations.
type SyncResult struct {
	DocCounts     map[string]int
	FetchDuration time.Duration
	PushDuration  time.Duration
}

// SyncError wraps a sync error with the source that caused it.
type SyncError struct {
	Source string // "netbox" or "glean"
	Err    error
}

func (e *SyncError) Error() string { return e.Err.Error() }
func (e *SyncError) Unwrap() error { return e.Err }

// SyncAll syncs all specified object types. If objectTypes is empty, syncs all registered crawlers.
// If since is non-nil, only objects updated after that time are fetched.
func (s *Syncer) SyncAll(ctx context.Context, objectTypes []string, since *time.Time) (*SyncResult, error) {
	crawlers := s.resolveCrawlers(objectTypes)

	slog.Info("starting sync", "objectTypes", len(crawlers), "incremental", since != nil)

	// Phase 1: Fetch all data from NetBox concurrently.
	type fetchResult struct {
		crawler crawler.Crawler
		docs    []components.DocumentDefinition
	}
	results := make([]fetchResult, len(crawlers))

	fetchStart := time.Now()
	g, fetchCtx := errgroup.WithContext(ctx)
	g.SetLimit(s.Concurrency)

	for i, c := range crawlers {
		g.Go(func() error {
			slog.Info("fetching from NetBox", "type", c.ObjectType(), "endpoint", c.Endpoint())
			params := url.Values{}
			if since != nil {
				params.Set("last_updated__gte", since.Format(time.RFC3339))
			}

			var docs []components.DocumentDefinition
			err := s.NetBox.List(fetchCtx, c.Endpoint(), params, func(objs []map[string]any) error {
				for _, obj := range objs {
					docs = append(docs, c.Transform(obj, s.Datasource, s.NetBoxURL))
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("fetching %s from NetBox: %w", c.ObjectType(), err)
			}
			results[i] = fetchResult{crawler: c, docs: docs}
			slog.Info("fetched from NetBox", "type", c.ObjectType(), "documents", len(docs))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, &SyncError{Source: "netbox", Err: fmt.Errorf("fetch phase failed: %w", err)}
	}
	fetchDuration := time.Since(fetchStart)

	// Phase 2: Push to Glean sequentially (Glean allows only one bulk upload at a time per datasource).
	pushStart := time.Now()
	docCounts := make(map[string]int, len(results))
	for _, r := range results {
		if err := s.pushToGlean(ctx, r.crawler, r.docs); err != nil {
			return nil, &SyncError{Source: "glean", Err: fmt.Errorf("push phase failed: %w", err)}
		}
		docCounts[r.crawler.ObjectType()] = len(r.docs)
	}
	pushDuration := time.Since(pushStart)

	// Phase 3: Tell Glean to process all uploaded documents.
	slog.Info("triggering document processing")
	ds := s.Datasource
	if _, err := s.GleanSDK.Indexing.Documents.ProcessAll(ctx, &components.ProcessAllDocumentsRequest{
		Datasource: &ds,
	}); err != nil {
		return nil, &SyncError{Source: "glean", Err: fmt.Errorf("process all documents: %w", err)}
	}

	slog.Info("sync completed successfully")
	return &SyncResult{
		DocCounts:     docCounts,
		FetchDuration: fetchDuration,
		PushDuration:  pushDuration,
	}, nil
}

func (s *Syncer) resolveCrawlers(objectTypes []string) []crawler.Crawler {
	all := crawler.All()
	if len(objectTypes) == 0 {
		result := make([]crawler.Crawler, 0, len(all))
		names := make([]string, 0, len(all))
		for name := range all {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			result = append(result, all[name])
		}
		return result
	}

	result := make([]crawler.Crawler, 0, len(objectTypes))
	for _, ot := range objectTypes {
		if c, ok := all[ot]; ok {
			result = append(result, c)
		} else {
			slog.Warn("unknown object type, skipping", "objectType", ot)
		}
	}
	return result
}

func (s *Syncer) pushToGlean(ctx context.Context, c crawler.Crawler, docs []components.DocumentDefinition) error {
	uploadID := fmt.Sprintf("%s-%s-%d", s.Datasource, c.ObjectType(), time.Now().UnixNano())

	if len(docs) == 0 {
		slog.Info("no documents to push", "type", c.ObjectType())
		_, err := s.GleanSDK.Indexing.Documents.BulkIndex(ctx, components.BulkIndexDocumentsRequest{
			UploadID:           uploadID,
			IsFirstPage:        crawler.Ptr(true),
			IsLastPage:         crawler.Ptr(true),
			ForceRestartUpload: crawler.Ptr(true),
			Datasource:         s.Datasource,
			Documents:          []components.DocumentDefinition{},
		})
		return err
	}

	slog.Info("pushing to Glean", "type", c.ObjectType(), "documents", len(docs), "uploadID", uploadID)

	isFirst := true
	for i := 0; i < len(docs); i += s.BatchSize {
		end := i + s.BatchSize
		if end > len(docs) {
			end = len(docs)
		}
		batch := docs[i:end]
		isLast := end >= len(docs)

		req := components.BulkIndexDocumentsRequest{
			UploadID:   uploadID,
			IsLastPage: crawler.Ptr(isLast),
			Datasource: s.Datasource,
			Documents:  batch,
		}
		if isFirst {
			req.IsFirstPage = crawler.Ptr(true)
			req.ForceRestartUpload = crawler.Ptr(true)
			isFirst = false
		}

		slog.Debug("sending bulk index batch", "type", c.ObjectType(), "docs", len(batch), "isLast", isLast)
		if _, err := s.GleanSDK.Indexing.Documents.BulkIndex(ctx, req); err != nil {
			return fmt.Errorf("bulk indexing %s: %w", c.ObjectType(), err)
		}
	}

	slog.Info("pushed to Glean", "type", c.ObjectType(), "documents", len(docs))
	return nil
}

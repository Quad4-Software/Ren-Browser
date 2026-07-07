// SPDX-License-Identifier: MIT
package app

import (
	"testing"
	"time"

	"renbrowser/internal/cache"
	"renbrowser/internal/nomadnet"
)

func TestPageResponseFromCacheMarksMetadata(t *testing.T) {
	svc := newTestBrowserService(t)
	stored := time.Now().Add(-30 * time.Minute)
	entry := cache.Entry{
		Body:        []byte("`=Title\nHello"),
		ContentType: "micron",
		StoredAt:    stored,
	}
	parsed := nomadnet.PageURL{NodeHash: "abb3ebcd03cb2388a838e70c001291f9", Path: "/page/index.mu"}
	resp := svc.pageResponseFromCache("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu", parsed, entry)
	if !resp.FromCache {
		t.Fatal("expected fromCache")
	}
	if resp.CachedAt != stored.UnixMilli() {
		t.Fatalf("cachedAt=%d want %d", resp.CachedAt, stored.UnixMilli())
	}
	if resp.HTML == "" {
		t.Fatal("expected rendered html")
	}
}

func TestStaleCacheEntryDetected(t *testing.T) {
	entry := cache.Entry{StoredAt: time.Now().Add(-cache.DefaultPageCacheMaxAge - time.Minute)}
	if !entry.IsStale(cache.DefaultPageCacheMaxAge) {
		t.Fatal("expected stale entry")
	}
}

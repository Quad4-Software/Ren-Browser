// SPDX-License-Identifier: MIT
package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"renbrowser/internal/nomadnet"
)

func TestPageCacheDiskPersistsAcrossReopen(t *testing.T) {
	dir := t.TempDir()
	req := nomadnet.RequestData{}
	c, err := OpenPageCache(dir, PageCacheOptions{
		RAMMaxEntries:  8,
		RAMMaxBytes:    1 << 20,
		DiskMaxEntries: 32,
		DiskMaxBytes:   4 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	c.Put("node", "/page/index.mu", req, []byte("hello-disk"), "micron")
	if c.Len() != 1 {
		t.Fatalf("len = %d", c.Len())
	}

	c2, err := OpenPageCache(dir, PageCacheOptions{
		RAMMaxEntries:  8,
		RAMMaxBytes:    1 << 20,
		DiskMaxEntries: 32,
		DiskMaxBytes:   4 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := c2.Get("node", "/page/index.mu", req)
	if !ok {
		t.Fatal("expected disk hit after reopen")
	}
	if string(entry.Body) != "hello-disk" {
		t.Fatalf("body = %q", entry.Body)
	}
	if entry.ContentType != "micron" {
		t.Fatalf("contentType = %q", entry.ContentType)
	}
}

func TestPageCacheGetPromotesDiskToRAM(t *testing.T) {
	dir := t.TempDir()
	req := nomadnet.RequestData{}
	c, err := OpenPageCache(dir, PageCacheOptions{
		RAMMaxEntries:  8,
		RAMMaxBytes:    1 << 20,
		DiskMaxEntries: 32,
		DiskMaxBytes:   4 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	c.Put("node", "/page/index.mu", req, []byte("promote-me"), "micron")
	stats := c.Stats()
	if stats.RAMEntries != 1 || stats.DiskEntries != 1 {
		t.Fatalf("stats before clear ram = %+v", stats)
	}

	c.mu.Lock()
	c.entries = make(map[pageKey]Entry, c.ramMax)
	c.order = c.order[:0]
	c.ramBytes = 0
	c.mu.Unlock()

	if stats := c.Stats(); stats.RAMEntries != 0 || stats.DiskEntries != 1 {
		t.Fatalf("expected disk-only state, got %+v", stats)
	}
	entry, ok := c.Get("node", "/page/index.mu", req)
	if !ok || string(entry.Body) != "promote-me" {
		t.Fatalf("promote get = %#v ok=%v", entry, ok)
	}
	if stats := c.Stats(); stats.RAMEntries != 1 {
		t.Fatalf("expected RAM promotion, got %+v", stats)
	}
}

func TestPageCacheClearRemovesDiskFiles(t *testing.T) {
	dir := t.TempDir()
	req := nomadnet.RequestData{}
	c, err := OpenPageCache(dir, PageCacheOptions{
		RAMMaxEntries:  8,
		RAMMaxBytes:    1 << 20,
		DiskMaxEntries: 32,
		DiskMaxBytes:   4 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	c.Put("node", "/page/a.mu", req, []byte("a"), "micron")
	c.Put("node", "/page/b.mu", req, []byte("b"), "micron")
	objectsDir := filepath.Join(dir, "objects")
	before, err := os.ReadDir(objectsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) == 0 {
		t.Fatal("expected disk objects before clear")
	}
	cleared := c.Clear()
	if cleared != 2 {
		t.Fatalf("cleared = %d", cleared)
	}
	after, err := os.ReadDir(objectsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 0 {
		t.Fatalf("expected empty objects dir, got %d files", len(after))
	}
	if _, ok := c.Get("node", "/page/a.mu", req); ok {
		t.Fatal("expected miss after clear")
	}
}

func TestPageCacheDiskByteBudgetEvictsOldest(t *testing.T) {
	dir := t.TempDir()
	req := nomadnet.RequestData{}
	c, err := OpenPageCache(dir, PageCacheOptions{
		RAMMaxEntries:  64,
		RAMMaxBytes:    1 << 20,
		DiskMaxEntries: 32,
		DiskMaxBytes:   1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	body := make([]byte, 400)
	for i := range 5 {
		path := "/page/" + string(byte('a'+i)) + ".mu"
		c.Put("node", path, req, body, "micron")
	}
	stats := c.Stats()
	if stats.DiskBytes > stats.DiskMaxBytes {
		t.Fatalf("disk bytes = %d max = %d", stats.DiskBytes, stats.DiskMaxBytes)
	}
	if stats.DiskEntries > 2 {
		t.Fatalf("expected disk byte budget to keep at most 2 entries, got %d", stats.DiskEntries)
	}
	c.mu.Lock()
	c.entries = make(map[pageKey]Entry, c.ramMax)
	c.order = c.order[:0]
	c.ramBytes = 0
	c.mu.Unlock()
	if _, ok := c.Get("node", "/page/a.mu", req); ok {
		t.Fatal("oldest disk entry should be evicted")
	}
	if _, ok := c.Get("node", "/page/e.mu", req); !ok {
		t.Fatal("newest disk entry should remain")
	}
}

func TestPageCacheStaleStillReturned(t *testing.T) {
	dir := t.TempDir()
	req := nomadnet.RequestData{}
	c, err := OpenPageCache(dir, PageCacheOptions{
		RAMMaxEntries:  8,
		RAMMaxBytes:    1 << 20,
		DiskMaxEntries: 32,
		DiskMaxBytes:   4 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	c.Put("node", "/page/index.mu", req, []byte("old"), "micron")
	c.mu.Lock()
	key := pageCacheKey("node", "/page/index.mu", req)
	entry := c.entries[key]
	entry.StoredAt = time.Now().Add(-48 * time.Hour)
	c.entries[key] = entry
	hash := pageKeyHash(key)
	meta := c.diskIndex[hash]
	meta.StoredAtUnixMilli = entry.StoredAt.UnixMilli()
	c.diskIndex[hash] = meta
	_ = c.writeDiskObject(hash, entry.Body, meta)
	c.mu.Unlock()

	got, ok := c.Get("node", "/page/index.mu", req)
	if !ok {
		t.Fatal("expected cache hit for stale entry")
	}
	if !got.IsStale(DefaultPageCacheMaxAge) {
		t.Fatal("expected entry to be stale")
	}
}

func TestOpenPageCacheEmptyDirIsRAMOnly(t *testing.T) {
	c, err := OpenPageCache("", PageCacheOptions{RAMMaxEntries: 4, RAMMaxBytes: 1024})
	if err != nil {
		t.Fatal(err)
	}
	if c.diskEnabled() {
		t.Fatal("expected RAM-only cache")
	}
	req := nomadnet.RequestData{}
	c.Put("node", "/page/index.mu", req, []byte("ram"), "micron")
	if c.Stats().DiskEntries != 0 {
		t.Fatal("expected no disk entries")
	}
}

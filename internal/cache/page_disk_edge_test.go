// SPDX-License-Identifier: MIT
package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"renbrowser/internal/nomadnet"
)

func TestPageCachePutUnwritableDiskFallsBackToRAM(t *testing.T) {
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
	objects := filepath.Join(dir, "objects")
	if err := os.Chmod(objects, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(objects, 0o700) })

	c.Put("node", "/page/index.mu", req, []byte("still-cached"), "micron")
	entry, ok := c.Get("node", "/page/index.mu", req)
	if !ok || string(entry.Body) != "still-cached" {
		t.Fatalf("expected RAM hit after disk write failure, got %#v ok=%v", entry, ok)
	}
	stats := c.Stats()
	if stats.DiskEntries != 0 {
		t.Fatalf("expected no disk entries, got %d", stats.DiskEntries)
	}
	if c.LastDiskError() == "" {
		t.Fatal("expected LastDiskError after unwritable disk")
	}
}

func TestPageCachePutDiskFullFallsBackToRAM(t *testing.T) {
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

	prev := writeTempDiskFile
	t.Cleanup(func() { writeTempDiskFile = prev })
	writeTempDiskFile = func(dir, pattern string, data []byte, perm os.FileMode) (string, error) {
		return "", &os.PathError{Op: "write", Path: filepath.Join(dir, pattern), Err: syscall.ENOSPC}
	}

	c.Put("node", "/page/index.mu", req, []byte("enospc"), "micron")
	entry, ok := c.Get("node", "/page/index.mu", req)
	if !ok || string(entry.Body) != "enospc" {
		t.Fatalf("expected RAM hit on ENOSPC, got %#v ok=%v", entry, ok)
	}
	if c.Stats().DiskEntries != 0 {
		t.Fatal("expected disk spill to stay empty on ENOSPC")
	}
	if !errors.Is(c.diskErr, syscall.ENOSPC) {
		t.Fatalf("expected ENOSPC, got %v", c.diskErr)
	}
}

func TestPageCachePutPartialCommitCleansIndex(t *testing.T) {
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
	prevRename := renameDiskFile
	t.Cleanup(func() { renameDiskFile = prevRename })
	failMetaRename := false
	renameDiskFile = func(oldpath, newpath string) error {
		if failMetaRename && filepath.Ext(newpath) == ".meta" {
			return &os.PathError{Op: "rename", Path: newpath, Err: syscall.ENOSPC}
		}
		return prevRename(oldpath, newpath)
	}

	c.Put("node", "/page/index.mu", req, []byte("v1"), "micron")
	if c.Stats().DiskEntries != 1 {
		t.Fatal("expected initial disk entry")
	}
	failMetaRename = true
	c.Put("node", "/page/index.mu", req, []byte("v2"), "micron")
	if _, ok := c.Get("node", "/page/index.mu", req); !ok {
		t.Fatal("expected RAM hit after partial disk failure")
	}
	if !errors.Is(c.diskErr, errDiskPartialCommit) {
		t.Fatalf("expected partial commit error, got %v", c.diskErr)
	}

	c.mu.Lock()
	c.entries = make(map[pageKey]Entry, c.ramMax)
	c.order = c.order[:0]
	c.ramBytes = 0
	c.mu.Unlock()

	if _, ok := c.Get("node", "/page/index.mu", req); ok {
		t.Fatal("partial commit should not leave a readable disk entry")
	}
	if c.Stats().DiskEntries != 0 {
		t.Fatalf("expected disk index cleared, got %d", c.Stats().DiskEntries)
	}
}

func TestOpenPageCacheUnwritableRootFallsBackToRAM(t *testing.T) {
	parent := t.TempDir()
	if err := os.Chmod(parent, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })

	c, err := OpenPageCache(filepath.Join(parent, "page-cache"), PageCacheOptions{
		RAMMaxEntries: 4,
		RAMMaxBytes:   1024,
	})
	if err != nil {
		t.Fatalf("OpenPageCache should not fail: %v", err)
	}
	if c.DiskEnabled() {
		t.Fatal("expected RAM-only fallback when mkdir fails")
	}
	if c.LastDiskError() == "" {
		t.Fatal("expected LastDiskError when mkdir fails")
	}
	req := nomadnet.RequestData{}
	c.Put("node", "/page/index.mu", req, []byte("ok"), "micron")
	if _, ok := c.Get("node", "/page/index.mu", req); !ok {
		t.Fatal("expected RAM cache to work after open fallback")
	}
}

func TestPageCacheGetMissingBinDropsIndexEntry(t *testing.T) {
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
	c.Put("node", "/page/index.mu", req, []byte("gone"), "micron")
	key := pageCacheKey("node", "/page/index.mu", req)
	hash := pageKeyHash(key)

	c.mu.Lock()
	c.entries = make(map[pageKey]Entry, c.ramMax)
	c.order = c.order[:0]
	c.ramBytes = 0
	c.mu.Unlock()

	if err := os.Remove(c.objectPath(hash, ".bin")); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Get("node", "/page/index.mu", req); ok {
		t.Fatal("expected miss when bin is missing")
	}
	if c.Stats().DiskEntries != 0 {
		t.Fatal("expected missing bin to drop disk index entry")
	}
	if c.LastDiskError() == "" {
		t.Fatal("expected LastDiskError for missing bin")
	}
}

func TestPageCacheLoadSkipsCorruptMetaAndOrphans(t *testing.T) {
	dir := t.TempDir()
	objects := filepath.Join(dir, "objects")
	if err := os.MkdirAll(objects, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(objects, "bad.meta"), []byte("{not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(objects, "orphan.bin"), []byte("orphan"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(objects, "leftover.tmp"), []byte("tmp"), 0o600); err != nil {
		t.Fatal(err)
	}

	good := diskMeta{
		Node:              "node",
		Path:              "/page/ok.mu",
		ContentType:       "micron",
		StoredAtUnixMilli: 1,
		Size:              5,
	}
	raw, err := json.Marshal(good)
	if err != nil {
		t.Fatal(err)
	}
	hash := pageKeyHash(pageKey{node: "node", path: "/page/ok.mu"})
	if err := os.WriteFile(filepath.Join(objects, hash+".bin"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(objects, hash+".meta"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	c, err := OpenPageCache(dir, PageCacheOptions{
		RAMMaxEntries:  8,
		RAMMaxBytes:    1 << 20,
		DiskMaxEntries: 32,
		DiskMaxBytes:   4 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Stats().DiskEntries != 1 {
		t.Fatalf("expected only good entry, got %d", c.Stats().DiskEntries)
	}
	if _, err := os.Stat(filepath.Join(objects, "bad.meta")); !os.IsNotExist(err) {
		t.Fatal("expected corrupt meta removed")
	}
	if _, err := os.Stat(filepath.Join(objects, "orphan.bin")); !os.IsNotExist(err) {
		t.Fatal("expected orphan bin removed")
	}
	if _, err := os.Stat(filepath.Join(objects, "leftover.tmp")); !os.IsNotExist(err) {
		t.Fatal("expected tmp cleaned on load")
	}
	entry, ok := c.Get("node", "/page/ok.mu", nomadnet.RequestData{})
	if !ok || string(entry.Body) != "hello" {
		t.Fatalf("expected good entry, got %#v ok=%v", entry, ok)
	}
}

func TestPageCacheSkipsOversizedDiskWrite(t *testing.T) {
	dir := t.TempDir()
	req := nomadnet.RequestData{}
	c, err := OpenPageCache(dir, PageCacheOptions{
		RAMMaxEntries:  8,
		RAMMaxBytes:    1 << 20,
		DiskMaxEntries: 32,
		DiskMaxBytes:   64,
	})
	if err != nil {
		t.Fatal(err)
	}
	body := make([]byte, 128)
	c.Put("node", "/page/big.mu", req, body, "micron")
	if c.Stats().DiskEntries != 0 {
		t.Fatal("expected oversized body to skip disk")
	}
	entry, ok := c.Get("node", "/page/big.mu", req)
	if !ok || len(entry.Body) != 128 {
		t.Fatal("expected oversized body retained in RAM")
	}
}

func TestPageCacheClearWithUnwritableDiskClearsRAM(t *testing.T) {
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
	objects := filepath.Join(dir, "objects")
	if err := os.Chmod(objects, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(objects, 0o700) })

	cleared := c.Clear()
	if cleared != 1 {
		t.Fatalf("cleared = %d", cleared)
	}
	if c.Len() != 0 || c.Stats().RAMEntries != 0 || c.Stats().DiskEntries != 0 {
		t.Fatalf("expected empty cache after clear, got %+v", c.Stats())
	}
	if _, ok := c.Get("node", "/page/a.mu", req); ok {
		t.Fatal("expected miss after clear with unwritable disk")
	}
}

func TestPageCacheDiskWriteRecoversAfterError(t *testing.T) {
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

	prev := writeTempDiskFile
	t.Cleanup(func() { writeTempDiskFile = prev })
	writeTempDiskFile = func(dir, pattern string, data []byte, perm os.FileMode) (string, error) {
		return "", &os.PathError{Op: "write", Path: pattern, Err: syscall.ENOSPC}
	}
	c.Put("node", "/page/index.mu", req, []byte("fail"), "micron")
	if c.LastDiskError() == "" {
		t.Fatal("expected disk error")
	}

	writeTempDiskFile = prev
	c.Put("node", "/page/index.mu", req, []byte("ok"), "micron")
	if c.LastDiskError() != "" {
		t.Fatalf("expected disk error cleared after success, got %q", c.LastDiskError())
	}
	if c.Stats().DiskEntries != 1 {
		t.Fatal("expected disk write to recover")
	}

	c.mu.Lock()
	c.entries = make(map[pageKey]Entry, c.ramMax)
	c.order = c.order[:0]
	c.ramBytes = 0
	c.mu.Unlock()
	entry, ok := c.Get("node", "/page/index.mu", req)
	if !ok || string(entry.Body) != "ok" {
		t.Fatalf("expected disk hit after recovery, got %#v ok=%v", entry, ok)
	}
}

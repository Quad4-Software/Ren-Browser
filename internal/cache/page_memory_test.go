// SPDX-License-Identifier: MIT
package cache

import (
	"fmt"
	"runtime"
	"testing"

	"renbrowser/internal/nomadnet"
)

func TestPageCacheRAMBudgetBoundsHeapGrowth(t *testing.T) {
	const (
		ramMaxBytes = 4 << 20
		pageSize    = 256 << 10
		puts        = 200
	)
	c := NewPageCacheWithBudget(1024, ramMaxBytes)
	req := nomadnet.RequestData{}
	body := make([]byte, pageSize)
	for i := range body {
		body[i] = byte(i)
	}

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	for i := range puts {
		path := fmt.Sprintf("/page/%d.mu", i)
		c.Put("node", path, req, body, "micron")
		if c.Bytes() > ramMaxBytes {
			t.Fatalf("RAM bytes %d exceeded budget %d after put %d", c.Bytes(), ramMaxBytes, i)
		}
	}

	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	stats := c.Stats()
	uncapped := int64(puts) * int64(pageSize)
	t.Logf("puts=%d page=%dKiB uncapped_body=%dMiB", puts, pageSize>>10, uncapped>>20)
	t.Logf("cache RAM entries=%d bytes=%d maxBytes=%d", stats.RAMEntries, stats.Bytes, stats.MaxBytes)
	t.Logf("heap_inuse before=%dMiB after=%dMiB delta=%dMiB",
		before.HeapInuse>>20, after.HeapInuse>>20, int64(after.HeapInuse-before.HeapInuse)>>20)

	if stats.Bytes > ramMaxBytes {
		t.Fatalf("cache bytes %d > budget %d", stats.Bytes, ramMaxBytes)
	}
	if stats.RAMEntries > (ramMaxBytes/pageSize)+1 {
		t.Fatalf("too many RAM entries: %d", stats.RAMEntries)
	}
	// Heap should stay near the budget, not near uncapped body size.
	if after.HeapInuse > before.HeapInuse+uint64(ramMaxBytes)*3+8<<20 {
		t.Fatalf("heap grew too much: before=%d after=%d budget=%d", before.HeapInuse, after.HeapInuse, ramMaxBytes)
	}
}

func TestPageCacheDiskSpillKeepsRAMBounded(t *testing.T) {
	const (
		ramMaxBytes  = 1 << 20
		diskMaxBytes = 8 << 20
		pageSize     = 128 << 10
		puts         = 120
	)
	dir := t.TempDir()
	c, err := OpenPageCache(dir, PageCacheOptions{
		RAMMaxEntries:  256,
		RAMMaxBytes:    ramMaxBytes,
		DiskMaxEntries: 512,
		DiskMaxBytes:   diskMaxBytes,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := nomadnet.RequestData{}
	body := make([]byte, pageSize)
	for i := range body {
		body[i] = byte(i * 3)
	}

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	for i := range puts {
		path := fmt.Sprintf("/page/%d.mu", i)
		c.Put("node", path, req, body, "micron")
		if c.Bytes() > ramMaxBytes {
			t.Fatalf("RAM bytes %d exceeded budget %d after put %d", c.Bytes(), ramMaxBytes, i)
		}
	}

	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	stats := c.Stats()
	t.Logf("puts=%d page=%dKiB", puts, pageSize>>10)
	t.Logf("RAM entries=%d bytes=%d/%d", stats.RAMEntries, stats.RAMBytes, ramMaxBytes)
	t.Logf("disk entries=%d bytes=%d/%d", stats.DiskEntries, stats.DiskBytes, diskMaxBytes)
	t.Logf("heap_inuse before=%dMiB after=%dMiB delta=%dMiB",
		before.HeapInuse>>20, after.HeapInuse>>20, int64(after.HeapInuse-before.HeapInuse)>>20)

	if stats.RAMBytes > ramMaxBytes {
		t.Fatalf("RAM bytes %d > budget %d", stats.RAMBytes, ramMaxBytes)
	}
	if stats.DiskBytes > diskMaxBytes {
		t.Fatalf("disk bytes %d > budget %d", stats.DiskBytes, diskMaxBytes)
	}
	if stats.DiskEntries == 0 {
		t.Fatal("expected disk spill for overflow pages")
	}
	if after.HeapInuse > before.HeapInuse+uint64(ramMaxBytes)*3+8<<20 {
		t.Fatalf("heap grew too much with disk spill: before=%d after=%d", before.HeapInuse, after.HeapInuse)
	}
}

func TestPageCacheClearReleasesTrackedBytes(t *testing.T) {
	c := NewPageCacheWithBudget(64, 2<<20)
	req := nomadnet.RequestData{}
	body := make([]byte, 64<<10)
	for i := range 40 {
		c.Put("node", fmt.Sprintf("/page/%d.mu", i), req, body, "micron")
	}
	if c.Bytes() == 0 {
		t.Fatal("expected cached bytes before clear")
	}
	c.Clear()
	if c.Bytes() != 0 || c.Len() != 0 {
		t.Fatalf("expected empty cache after clear, bytes=%d len=%d", c.Bytes(), c.Len())
	}

	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	t.Logf("after clear heap_inuse=%dMiB heap_alloc=%dMiB", ms.HeapInuse>>20, ms.HeapAlloc>>20)
}

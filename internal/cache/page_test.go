// SPDX-License-Identifier: MIT
package cache

import (
	"bytes"
	"testing"

	"renbrowser/internal/nomadnet"
)

func TestPageCacheSkipsUnchangedPut(t *testing.T) {
	c := NewPageCache(4)
	req := nomadnet.RequestData{}
	body := []byte("same")
	c.Put("node", "/page/index.mu", req, body, "micron")
	first, ok := c.Get("node", "/page/index.mu", req)
	if !ok {
		t.Fatal("expected cache hit")
	}
	c.Put("node", "/page/index.mu", req, body, "micron")
	second, ok := c.Get("node", "/page/index.mu", req)
	if !ok {
		t.Fatal("expected cache hit after unchanged put")
	}
	if !bytes.Equal(first.Body, second.Body) || &first.Body[0] != &second.Body[0] {
		t.Fatal("expected unchanged put to reuse body buffer")
	}
}

func TestPageCacheClear(t *testing.T) {
	c := NewPageCache(4)
	req := nomadnet.RequestData{}
	c.Put("node", "/page/index.mu", req, []byte("one"), "micron")
	c.Put("node", "/page/other.mu", req, []byte("two"), "micron")
	if c.Len() != 2 {
		t.Fatalf("len = %d", c.Len())
	}
	cleared := c.Clear()
	if cleared != 2 {
		t.Fatalf("cleared = %d", cleared)
	}
	if c.Len() != 0 {
		t.Fatalf("len after clear = %d", c.Len())
	}
}

func TestPageCacheRAMLRUTouchOnGet(t *testing.T) {
	c := NewPageCacheWithBudget(2, 1<<20)
	req := nomadnet.RequestData{}
	c.Put("node", "/page/a.mu", req, []byte("a"), "micron")
	c.Put("node", "/page/b.mu", req, []byte("b"), "micron")
	if _, ok := c.Get("node", "/page/a.mu", req); !ok {
		t.Fatal("expected hit for a")
	}
	c.Put("node", "/page/c.mu", req, []byte("c"), "micron")
	if _, ok := c.Get("node", "/page/a.mu", req); !ok {
		t.Fatal("a should remain after LRU touch")
	}
	if _, ok := c.Get("node", "/page/b.mu", req); ok {
		t.Fatal("b should be evicted")
	}
}

func TestPageCacheByteBudgetEvictsOldest(t *testing.T) {
	c := NewPageCacheWithBudget(32, 1000)
	req := nomadnet.RequestData{}
	body := make([]byte, 400)
	for i := range 5 {
		path := "/page/" + string(byte('a'+i)) + ".mu"
		c.Put("node", path, req, body, "micron")
	}
	if c.Bytes() > c.MaxBytes() {
		t.Fatalf("bytes = %d maxBytes = %d", c.Bytes(), c.MaxBytes())
	}
	if c.Len() > 2 {
		t.Fatalf("expected byte budget to keep at most 2 entries, got %d", c.Len())
	}
}

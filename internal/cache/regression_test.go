// SPDX-License-Identifier: MIT
package cache_test

import (
	"strings"
	"testing"

	"renbrowser/internal/cache"
	"renbrowser/internal/nomadnet"
)

func TestPageCacheEvictsOldestEntry(t *testing.T) {
	c := cache.NewPageCache(2)
	req := nomadnet.RequestData{}
	c.Put("node", "/page/one.mu", req, []byte("1"), "micron")
	c.Put("node", "/page/two.mu", req, []byte("2"), "micron")
	c.Put("node", "/page/three.mu", req, []byte("3"), "micron")

	if _, ok := c.Get("node", "/page/one.mu", req); ok {
		t.Fatal("oldest entry should be evicted")
	}
	if _, ok := c.Get("node", "/page/two.mu", req); !ok {
		t.Fatal("expected middle entry to remain")
	}
	if _, ok := c.Get("node", "/page/three.mu", req); !ok {
		t.Fatal("expected newest entry to remain")
	}
	if c.Len() != 2 {
		t.Fatalf("len = %d; want 2", c.Len())
	}
}

func TestPageCacheIsolatesRequestSuffixes(t *testing.T) {
	c := cache.NewPageCache(8)
	base := nomadnet.RequestData{}
	withA := nomadnet.RequestData{Vars: map[string]string{"tab": "a"}}
	withB := nomadnet.RequestData{Vars: map[string]string{"tab": "b"}}

	c.Put("node", "/page/form.mu", base, []byte("base"), "micron")
	c.Put("node", "/page/form.mu", withA, []byte("a"), "micron")
	c.Put("node", "/page/form.mu", withB, []byte("b"), "micron")

	entry, ok := c.Get("node", "/page/form.mu", withA)
	if !ok || string(entry.Body) != "a" {
		t.Fatalf("entry a = %#v ok=%v", entry, ok)
	}
	entry, ok = c.Get("node", "/page/form.mu", withB)
	if !ok || string(entry.Body) != "b" {
		t.Fatalf("entry b = %#v ok=%v", entry, ok)
	}
	entry, ok = c.Get("node", "/page/form.mu", base)
	if !ok || string(entry.Body) != "base" {
		t.Fatalf("entry base = %#v ok=%v", entry, ok)
	}
}

func TestPageCacheUpdatesWhenContentChanges(t *testing.T) {
	c := cache.NewPageCache(4)
	req := nomadnet.RequestData{}
	c.Put("node", "/page/index.mu", req, []byte("v1"), "micron")
	c.Put("node", "/page/index.mu", req, []byte("v2"), "micron")

	entry, ok := c.Get("node", "/page/index.mu", req)
	if !ok || string(entry.Body) != "v2" {
		t.Fatalf("body = %q ok=%v", entry.Body, ok)
	}
}

func TestPageCacheEvictionDoesNotGrowOrderWithoutBound(t *testing.T) {
	c := cache.NewPageCache(2)
	req := nomadnet.RequestData{}
	for i := range 32 {
		path := "/page/" + strings.Repeat("x", i%3) + ".mu"
		c.Put("node", path, req, []byte{byte(i)}, "micron")
	}
	if c.Len() != 2 {
		t.Fatalf("len = %d; want 2", c.Len())
	}
}

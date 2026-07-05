// SPDX-License-Identifier: MIT
package cache_test

import (
	"testing"

	"renbrowser/internal/cache"
	"renbrowser/internal/nomadnet"
)

func BenchmarkPageCachePutGet(b *testing.B) {
	c := cache.NewPageCache(128)
	body := []byte("sample page body for cache benchmark")
	node := "abb3ebcd03cb2388a838e70c001291f9"
	path := "/page/index.mu"
	empty := nomadnet.RequestData{}
	c.Put(node, path, empty, body, "micron")

	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	for b.Loop() {
		c.Put(node, path, empty, body, "micron")
		if _, ok := c.Get(node, path, empty); !ok {
			b.Fatal("cache miss")
		}
	}
}

func BenchmarkPageCachePutGetReuseBuffer(b *testing.B) {
	c := cache.NewPageCache(128)
	body := []byte("sample page body for cache benchmark")
	node := "abb3ebcd03cb2388a838e70c001291f9"
	path := "/page/index.mu"
	empty := nomadnet.RequestData{}
	c.Put(node, path, empty, body, "micron")

	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	for b.Loop() {
		c.Put(node, path, empty, body, "micron")
	}
}

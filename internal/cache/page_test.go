// SPDX-License-Identifier: MIT
package cache_test

import (
	"testing"

	"renbrowser/internal/cache"
	"renbrowser/internal/nomadnet"
)

func TestPageCacheEviction(t *testing.T) {
	c := cache.NewPageCache(2)
	empty := nomadnet.RequestData{}
	c.Put("aa", "/page/a.mu", empty, []byte("a"), "micron")
	c.Put("bb", "/page/b.mu", empty, []byte("b"), "micron")
	c.Put("cc", "/page/c.mu", empty, []byte("c"), "micron")

	if c.Len() != 2 {
		t.Fatalf("len = %d; want 2", c.Len())
	}
	if _, ok := c.Get("aa", "/page/a.mu", empty); ok {
		t.Fatal("oldest entry should be evicted")
	}
	if _, ok := c.Get("cc", "/page/c.mu", empty); !ok {
		t.Fatal("latest entry missing")
	}
}

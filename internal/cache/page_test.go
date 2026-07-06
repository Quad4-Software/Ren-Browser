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

// SPDX-License-Identifier: MIT
package cache

import (
	"testing"

	"renbrowser/internal/nomadnet"
)

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

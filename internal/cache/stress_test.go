//go:build stress

// SPDX-License-Identifier: MIT
package cache_test

import (
	"sync"
	"testing"

	"renbrowser/internal/cache"
	"renbrowser/internal/nomadnet"
)

func TestPageCacheConcurrentPutGet(t *testing.T) {
	c := cache.NewPageCache(64)
	req := nomadnet.RequestData{}
	body := []byte("shared-body")
	c.Put("node", "/page/index.mu", req, body, "micron")

	const workers = 16
	const rounds = 128
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for i := range rounds {
				path := "/page/index.mu"
				if i%2 == 0 {
					c.Put("node", path, req, body, "micron")
				}
				if _, ok := c.Get("node", path, req); !ok {
					t.Error("cache miss under concurrency")
				}
			}
		}()
	}
	wg.Wait()
}

// SPDX-License-Identifier: MIT
package cache

import (
	"bytes"
	"sync"
	"time"

	"renbrowser/internal/nomadnet"
)

type pageKey struct {
	node   string
	path   string
	suffix string
}

type Entry struct {
	Body        []byte
	ContentType string
	StoredAt    time.Time
}

type PageCache struct {
	mu      sync.RWMutex
	max     int
	entries map[pageKey]Entry
	order   []pageKey
}

func NewPageCache(max int) *PageCache {
	if max <= 0 {
		max = 128
	}
	return &PageCache{
		max:     max,
		entries: make(map[pageKey]Entry, max),
		order:   make([]pageKey, 0, max),
	}
}

func pageCacheKey(nodeHash, path string, req nomadnet.RequestData) pageKey {
	return pageKey{node: nodeHash, path: path, suffix: req.CacheKeySuffix()}
}

func (c *PageCache) Get(nodeHash, path string, req nomadnet.RequestData) (Entry, bool) {
	key := pageCacheKey(nodeHash, path, req)
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	return entry, ok
}

func (c *PageCache) Put(nodeHash, path string, req nomadnet.RequestData, body []byte, contentType string) {
	key := pageCacheKey(nodeHash, path, req)
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.entries[key]; ok {
		if existing.ContentType == contentType && bytes.Equal(existing.Body, body) {
			return
		}
		stored := cloneBody(existing.Body, body)
		c.entries[key] = Entry{
			Body:        stored,
			ContentType: contentType,
			StoredAt:    time.Now(),
		}
		return
	}

	c.order = append(c.order, key)
	if len(c.order) > c.max {
		oldest := c.order[0]
		copy(c.order, c.order[1:])
		c.order = c.order[:len(c.order)-1]
		delete(c.entries, oldest)
	}

	c.entries[key] = Entry{
		Body:        cloneBody(nil, body),
		ContentType: contentType,
		StoredAt:    time.Now(),
	}
}

func cloneBody(existing, body []byte) []byte {
	if cap(existing) >= len(body) {
		out := existing[:len(body)]
		copy(out, body)
		return out
	}
	return append([]byte(nil), body...)
}

func (c *PageCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *PageCache) Max() int {
	return c.max
}

func (c *PageCache) Clear() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := len(c.entries)
	c.entries = make(map[pageKey]Entry, c.max)
	c.order = make([]pageKey, 0, c.max)
	return count
}

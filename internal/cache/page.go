package cache

import (
	"sync"
	"time"

	"renbrowser/internal/nomadnet"
)

type Entry struct {
	Body        []byte
	ContentType string
	StoredAt    time.Time
}

type PageCache struct {
	mu      sync.RWMutex
	max     int
	entries map[string]Entry
	order   []string
}

func NewPageCache(max int) *PageCache {
	if max <= 0 {
		max = 128
	}
	return &PageCache{
		max:     max,
		entries: make(map[string]Entry, max),
		order:   make([]string, 0, max),
	}
}

func (c *PageCache) Key(nodeHash, path string, req nomadnet.RequestData) string {
	return nodeHash + "|" + path + req.CacheKeySuffix()
}

func (c *PageCache) Get(nodeHash, path string, req nomadnet.RequestData) (Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[c.Key(nodeHash, path, req)]
	return entry, ok
}

func (c *PageCache) Put(nodeHash, path string, req nomadnet.RequestData, body []byte, contentType string) {
	key := c.Key(nodeHash, path, req)
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.entries[key]; !exists {
		c.order = append(c.order, key)
		if len(c.order) > c.max {
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.entries, oldest)
		}
	}

	stored := cloneBody(c.entries[key].Body, body)
	c.entries[key] = Entry{
		Body:        stored,
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

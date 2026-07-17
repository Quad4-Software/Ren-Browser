// SPDX-License-Identifier: MIT
package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"renbrowser/internal/nomadnet"
)

var (
	errDiskPartialCommit = errors.New("page cache disk commit incomplete")
	writeTempDiskFile    = writeTempDiskFileImpl
	renameDiskFile       = os.Rename
)

type pageKey struct {
	node   string
	path   string
	suffix string
}

const (
	DefaultPageCacheMaxAge = 24 * time.Hour
	DefaultRAMMaxEntries   = 64
	DefaultRAMMaxBytes     = 16 << 20
	DefaultDiskMaxEntries  = 512
	DefaultDiskMaxBytes    = 256 << 20
)

// DefaultPageCacheMaxBytes is the soft heap budget for RAM-only caches.
const DefaultPageCacheMaxBytes = DefaultRAMMaxBytes

type Entry struct {
	Body        []byte
	ContentType string
	StoredAt    time.Time
}

func (e Entry) IsStale(maxAge time.Duration) bool {
	if maxAge <= 0 {
		return false
	}
	return time.Since(e.StoredAt) > maxAge
}

// PageCacheOptions configures RAM and optional disk tiers for OpenPageCache.
type PageCacheOptions struct {
	RAMMaxEntries  int
	RAMMaxBytes    int
	DiskMaxEntries int
	DiskMaxBytes   int
}

// PageCacheStats reports RAM and disk tier usage.
type PageCacheStats struct {
	Entries        int
	Max            int
	Bytes          int
	MaxBytes       int
	RAMEntries     int
	RAMBytes       int
	DiskEntries    int
	DiskBytes      int
	DiskMaxEntries int
	DiskMaxBytes   int
}

type diskMeta struct {
	Node              string `json:"node"`
	Path              string `json:"path"`
	Suffix            string `json:"suffix"`
	ContentType       string `json:"contentType"`
	StoredAtUnixMilli int64  `json:"storedAtUnixMilli"`
	Size              int    `json:"size"`
}

type PageCache struct {
	mu sync.Mutex

	ramMax      int
	ramMaxBytes int
	ramBytes    int
	entries     map[pageKey]Entry
	order       []pageKey

	dir            string
	diskMax        int
	diskMaxBytes   int
	diskBytes      int
	diskIndex      map[string]diskMeta
	diskOrder      []string
	diskOrderIndex map[string]int
	diskErr        error
}

// NewPageCache creates a RAM-only page cache.
func NewPageCache(max int) *PageCache {
	return NewPageCacheWithBudget(max, DefaultPageCacheMaxBytes)
}

// NewPageCacheWithBudget creates a RAM-only page cache with an entry and byte budget.
func NewPageCacheWithBudget(max, maxBytes int) *PageCache {
	if max <= 0 {
		max = DefaultRAMMaxEntries
	}
	if maxBytes <= 0 {
		maxBytes = DefaultPageCacheMaxBytes
	}
	return &PageCache{
		ramMax:      max,
		ramMaxBytes: maxBytes,
		entries:     make(map[pageKey]Entry, max),
		order:       make([]pageKey, 0, max),
	}
}

// OpenPageCache opens a two-tier page cache with disk spill under dir.
// If the disk tree cannot be created or indexed, it returns a usable RAM-only
// cache and records the failure in LastDiskError.
func OpenPageCache(dir string, opts PageCacheOptions) (*PageCache, error) {
	if dir == "" {
		return NewPageCacheWithBudget(opts.RAMMaxEntries, opts.RAMMaxBytes), nil
	}
	opts = normalizePageCacheOptions(opts)
	objectsDir := filepath.Join(dir, "objects")
	if err := os.MkdirAll(objectsDir, 0o700); err != nil {
		c := NewPageCacheWithBudget(opts.RAMMaxEntries, opts.RAMMaxBytes)
		c.diskErr = err
		return c, nil
	}
	c := &PageCache{
		ramMax:         opts.RAMMaxEntries,
		ramMaxBytes:    opts.RAMMaxBytes,
		entries:        make(map[pageKey]Entry, opts.RAMMaxEntries),
		order:          make([]pageKey, 0, opts.RAMMaxEntries),
		dir:            dir,
		diskMax:        opts.DiskMaxEntries,
		diskMaxBytes:   opts.DiskMaxBytes,
		diskIndex:      make(map[string]diskMeta, opts.DiskMaxEntries),
		diskOrder:      make([]string, 0, opts.DiskMaxEntries),
		diskOrderIndex: make(map[string]int, opts.DiskMaxEntries),
	}
	if err := c.loadDiskIndex(); err != nil {
		c.diskErr = err
		c.diskIndex = make(map[string]diskMeta, opts.DiskMaxEntries)
		c.diskOrder = make([]string, 0, opts.DiskMaxEntries)
		c.diskOrderIndex = make(map[string]int, opts.DiskMaxEntries)
		c.diskBytes = 0
		return c, nil
	}
	c.evictDiskOverflowLocked()
	return c, nil
}

func normalizePageCacheOptions(opts PageCacheOptions) PageCacheOptions {
	if opts.RAMMaxEntries <= 0 {
		opts.RAMMaxEntries = DefaultRAMMaxEntries
	}
	if opts.RAMMaxBytes <= 0 {
		opts.RAMMaxBytes = DefaultRAMMaxBytes
	}
	if opts.DiskMaxEntries <= 0 {
		opts.DiskMaxEntries = DefaultDiskMaxEntries
	}
	if opts.DiskMaxBytes <= 0 {
		opts.DiskMaxBytes = DefaultDiskMaxBytes
	}
	return opts
}

func pageCacheKey(nodeHash, path string, req nomadnet.RequestData) pageKey {
	return pageKey{node: nodeHash, path: path, suffix: req.CacheKeySuffix()}
}

func pageKeyHash(key pageKey) string {
	sum := sha256.Sum256([]byte(key.node + "\x00" + key.path + "\x00" + key.suffix))
	return hex.EncodeToString(sum[:])
}

func (c *PageCache) diskEnabled() bool {
	return c.dir != ""
}

// DiskEnabled reports whether a disk spill directory is configured.
func (c *PageCache) DiskEnabled() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.diskEnabled()
}

// LastDiskError returns the most recent disk I/O failure message, if any.
func (c *PageCache) LastDiskError() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.diskErr == nil {
		return ""
	}
	return c.diskErr.Error()
}

func (c *PageCache) Get(nodeHash, path string, req nomadnet.RequestData) (Entry, bool) {
	key := pageCacheKey(nodeHash, path, req)
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.entries[key]; ok {
		c.touchRAMLocked(key)
		if c.diskEnabled() {
			hash := pageKeyHash(key)
			if _, onDisk := c.diskIndex[hash]; onDisk {
				c.touchDiskLocked(hash)
			}
		}
		return entry, true
	}
	if !c.diskEnabled() {
		return Entry{}, false
	}
	hash := pageKeyHash(key)
	meta, ok := c.diskIndex[hash]
	if !ok {
		return Entry{}, false
	}
	body, err := os.ReadFile(c.objectPath(hash, ".bin"))
	if err != nil {
		c.diskErr = err
		c.removeDiskEntryLocked(hash)
		return Entry{}, false
	}
	if meta.Size != len(body) {
		c.diskBytes -= meta.Size
		meta.Size = len(body)
		c.diskBytes += meta.Size
		c.diskIndex[hash] = meta
	}
	entry := Entry{
		Body:        body,
		ContentType: meta.ContentType,
		StoredAt:    time.UnixMilli(meta.StoredAtUnixMilli),
	}
	c.touchDiskLocked(hash)
	c.putRAMLocked(key, entry)
	return entry, true
}

func (c *PageCache) Put(nodeHash, path string, req nomadnet.RequestData, body []byte, contentType string) {
	key := pageCacheKey(nodeHash, path, req)
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.entries[key]; ok {
		if existing.ContentType == contentType && bytes.Equal(existing.Body, body) {
			c.touchRAMLocked(key)
			if c.diskEnabled() {
				hash := pageKeyHash(key)
				if _, onDisk := c.diskIndex[hash]; onDisk {
					c.touchDiskLocked(hash)
					return
				}
			} else {
				return
			}
		}
	}

	storedAt := time.Now()
	stored := cloneBody(nil, body)
	entry := Entry{
		Body:        stored,
		ContentType: contentType,
		StoredAt:    storedAt,
	}
	if c.diskEnabled() {
		c.putDiskLocked(key, entry)
	}
	c.putRAMLocked(key, entry)
}

func (c *PageCache) putRAMLocked(key pageKey, entry Entry) {
	if existing, ok := c.entries[key]; ok {
		c.ramBytes -= len(existing.Body)
		stored := cloneBody(existing.Body, entry.Body)
		c.entries[key] = Entry{
			Body:        stored,
			ContentType: entry.ContentType,
			StoredAt:    entry.StoredAt,
		}
		c.ramBytes += len(stored)
		c.touchRAMLocked(key)
		c.evictRAMOverflowLocked()
		return
	}
	stored := cloneBody(nil, entry.Body)
	c.entries[key] = Entry{
		Body:        stored,
		ContentType: entry.ContentType,
		StoredAt:    entry.StoredAt,
	}
	c.ramBytes += len(stored)
	c.order = append(c.order, key)
	c.evictRAMOverflowLocked()
}

func (c *PageCache) touchRAMLocked(key pageKey) {
	for i, k := range c.order {
		if k != key {
			continue
		}
		copy(c.order[i:], c.order[i+1:])
		c.order = c.order[:len(c.order)-1]
		break
	}
	c.order = append(c.order, key)
}

func (c *PageCache) putDiskLocked(key pageKey, entry Entry) {
	if len(entry.Body) > c.diskMaxBytes {
		return
	}
	hash := pageKeyHash(key)
	meta := diskMeta{
		Node:              key.node,
		Path:              key.path,
		Suffix:            key.suffix,
		ContentType:       entry.ContentType,
		StoredAtUnixMilli: entry.StoredAt.UnixMilli(),
		Size:              len(entry.Body),
	}
	if err := c.writeDiskObject(hash, entry.Body, meta); err != nil {
		c.diskErr = err
		if errors.Is(err, errDiskPartialCommit) {
			c.removeDiskEntryLocked(hash)
		}
		return
	}
	c.diskErr = nil
	if prev, ok := c.diskIndex[hash]; ok {
		c.diskBytes -= prev.Size
		c.removeDiskOrderLocked(hash)
	}
	c.diskIndex[hash] = meta
	c.diskBytes += meta.Size
	c.diskOrder = append(c.diskOrder, hash)
	c.diskOrderIndex[hash] = len(c.diskOrder) - 1
	c.evictDiskOverflowLocked()
}

func (c *PageCache) evictRAMOverflowLocked() {
	for len(c.order) > c.ramMax || c.ramBytes > c.ramMaxBytes {
		if len(c.order) == 0 {
			c.ramBytes = 0
			return
		}
		oldest := c.order[0]
		c.order = c.order[1:]
		if entry, ok := c.entries[oldest]; ok {
			c.ramBytes -= len(entry.Body)
			delete(c.entries, oldest)
		}
	}
	if c.ramBytes < 0 {
		c.ramBytes = 0
	}
}

func (c *PageCache) evictDiskOverflowLocked() {
	if !c.diskEnabled() {
		return
	}
	for len(c.diskOrder) > c.diskMax || c.diskBytes > c.diskMaxBytes {
		if len(c.diskOrder) == 0 {
			c.diskBytes = 0
			return
		}
		oldest := c.diskOrder[0]
		c.removeDiskEntryLocked(oldest)
	}
	if c.diskBytes < 0 {
		c.diskBytes = 0
	}
}

func (c *PageCache) touchDiskLocked(hash string) {
	c.removeDiskOrderLocked(hash)
	c.diskOrder = append(c.diskOrder, hash)
	c.diskOrderIndex[hash] = len(c.diskOrder) - 1
}

func (c *PageCache) removeDiskOrderLocked(hash string) {
	idx, ok := c.diskOrderIndex[hash]
	if !ok {
		return
	}
	copy(c.diskOrder[idx:], c.diskOrder[idx+1:])
	c.diskOrder = c.diskOrder[:len(c.diskOrder)-1]
	delete(c.diskOrderIndex, hash)
	for i := idx; i < len(c.diskOrder); i++ {
		c.diskOrderIndex[c.diskOrder[i]] = i
	}
}

func (c *PageCache) removeDiskEntryLocked(hash string) {
	if meta, ok := c.diskIndex[hash]; ok {
		c.diskBytes -= meta.Size
		delete(c.diskIndex, hash)
	}
	c.removeDiskOrderLocked(hash)
	_ = os.Remove(c.objectPath(hash, ".bin"))
	_ = os.Remove(c.objectPath(hash, ".meta"))
}

func (c *PageCache) objectPath(hash, ext string) string {
	return filepath.Join(c.dir, "objects", hash+ext)
}

func (c *PageCache) writeDiskObject(hash string, body []byte, meta diskMeta) error {
	objectsDir := filepath.Join(c.dir, "objects")
	binPath := c.objectPath(hash, ".bin")
	metaPath := c.objectPath(hash, ".meta")
	raw, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	binTmp, err := writeTempDiskFile(objectsDir, hash+".bin.*.tmp", body, 0o600)
	if err != nil {
		return err
	}
	metaTmp, err := writeTempDiskFile(objectsDir, hash+".meta.*.tmp", raw, 0o600)
	if err != nil {
		_ = os.Remove(binTmp)
		return err
	}

	if err := renameDiskFile(binTmp, binPath); err != nil {
		_ = os.Remove(binTmp)
		_ = os.Remove(metaTmp)
		return err
	}
	if err := renameDiskFile(metaTmp, metaPath); err != nil {
		_ = os.Remove(metaTmp)
		_ = os.Remove(binPath)
		_ = os.Remove(metaPath)
		return fmt.Errorf("%w: %v", errDiskPartialCommit, err)
	}
	return nil
}

func writeTempDiskFileImpl(dir, pattern string, data []byte, perm os.FileMode) (string, error) {
	tmp, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return "", err
	}
	if err := tmp.Chmod(perm); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	ok = true
	return tmpName, nil
}

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := writeTempDiskFile(dir, filepath.Base(path)+".*.tmp", data, perm)
	if err != nil {
		return err
	}
	if err := renameDiskFile(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (c *PageCache) loadDiskIndex() error {
	objectsDir := filepath.Join(c.dir, "objects")
	entries, err := os.ReadDir(objectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	type loaded struct {
		hash string
		meta diskMeta
	}
	var items []loaded
	haveMeta := make(map[string]struct{})
	for _, ent := range entries {
		name := ent.Name()
		if ent.IsDir() {
			continue
		}
		if strings.HasSuffix(name, ".tmp") {
			_ = os.Remove(filepath.Join(objectsDir, name))
			continue
		}
		if filepath.Ext(name) != ".meta" {
			continue
		}
		hash := name[:len(name)-len(".meta")]
		raw, err := os.ReadFile(filepath.Join(objectsDir, name))
		if err != nil {
			continue
		}
		var meta diskMeta
		if err := json.Unmarshal(raw, &meta); err != nil {
			_ = os.Remove(filepath.Join(objectsDir, name))
			continue
		}
		binPath := c.objectPath(hash, ".bin")
		info, err := os.Stat(binPath)
		if err != nil {
			_ = os.Remove(filepath.Join(objectsDir, name))
			continue
		}
		size := int(info.Size())
		if meta.Size <= 0 {
			meta.Size = size
		} else if meta.Size != size {
			meta.Size = size
		}
		haveMeta[hash] = struct{}{}
		items = append(items, loaded{hash: hash, meta: meta})
	}
	for _, ent := range entries {
		name := ent.Name()
		if ent.IsDir() || filepath.Ext(name) != ".bin" {
			continue
		}
		hash := name[:len(name)-len(".bin")]
		if _, ok := haveMeta[hash]; ok {
			continue
		}
		_ = os.Remove(filepath.Join(objectsDir, name))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].meta.StoredAtUnixMilli < items[j].meta.StoredAtUnixMilli
	})
	for _, item := range items {
		c.diskIndex[item.hash] = item.meta
		c.diskBytes += item.meta.Size
		c.diskOrder = append(c.diskOrder, item.hash)
		c.diskOrderIndex[item.hash] = len(c.diskOrder) - 1
	}
	return nil
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
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.uniqueCountLocked()
}

func (c *PageCache) uniqueCountLocked() int {
	if !c.diskEnabled() {
		return len(c.entries)
	}
	seen := make(map[pageKey]struct{}, len(c.entries)+len(c.diskIndex))
	for key := range c.entries {
		seen[key] = struct{}{}
	}
	for _, meta := range c.diskIndex {
		seen[pageKey{node: meta.Node, path: meta.Path, suffix: meta.Suffix}] = struct{}{}
	}
	return len(seen)
}

func (c *PageCache) Bytes() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ramBytes
}

func (c *PageCache) Max() int {
	if c.diskEnabled() {
		return c.diskMax
	}
	return c.ramMax
}

func (c *PageCache) MaxBytes() int {
	if c.diskEnabled() {
		return c.diskMaxBytes
	}
	return c.ramMaxBytes
}

func (c *PageCache) Stats() PageCacheStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	stats := PageCacheStats{
		Entries:    c.uniqueCountLocked(),
		RAMEntries: len(c.entries),
		RAMBytes:   c.ramBytes,
	}
	if c.diskEnabled() {
		stats.Max = c.diskMax
		stats.MaxBytes = c.diskMaxBytes
		stats.Bytes = c.ramBytes + c.diskBytes
		stats.DiskEntries = len(c.diskIndex)
		stats.DiskBytes = c.diskBytes
		stats.DiskMaxEntries = c.diskMax
		stats.DiskMaxBytes = c.diskMaxBytes
	} else {
		stats.Max = c.ramMax
		stats.MaxBytes = c.ramMaxBytes
		stats.Bytes = c.ramBytes
	}
	return stats
}

func (c *PageCache) Clear() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := c.uniqueCountLocked()
	c.entries = make(map[pageKey]Entry, c.ramMax)
	c.order = make([]pageKey, 0, c.ramMax)
	c.ramBytes = 0
	if c.diskEnabled() {
		for hash := range c.diskIndex {
			_ = os.Remove(c.objectPath(hash, ".bin"))
			_ = os.Remove(c.objectPath(hash, ".meta"))
		}
		c.diskIndex = make(map[string]diskMeta, c.diskMax)
		c.diskOrder = make([]string, 0, c.diskMax)
		c.diskOrderIndex = make(map[string]int, c.diskMax)
		c.diskBytes = 0
		objectsDir := filepath.Join(c.dir, "objects")
		if entries, err := os.ReadDir(objectsDir); err == nil {
			for _, ent := range entries {
				_ = os.Remove(filepath.Join(objectsDir, ent.Name()))
			}
		}
	}
	return count
}

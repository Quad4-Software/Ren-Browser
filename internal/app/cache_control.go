// SPDX-License-Identifier: MIT
package app

import "renbrowser/internal/cache"

type PageCacheStats struct {
	Entries      int `json:"entries"`
	Max          int `json:"max"`
	Bytes        int `json:"bytes"`
	MaxBytes     int `json:"maxBytes"`
	RAMEntries   int `json:"ramEntries"`
	DiskEntries  int `json:"diskEntries"`
	DiskBytes    int `json:"diskBytes"`
	DiskMaxBytes int `json:"diskMaxBytes"`
}

func (s *BrowserService) GetPageCacheStats() PageCacheStats {
	if s == nil || s.pageCache == nil {
		return PageCacheStats{
			Max:          cache.DefaultDiskMaxEntries,
			MaxBytes:     cache.DefaultDiskMaxBytes,
			DiskMaxBytes: cache.DefaultDiskMaxBytes,
		}
	}
	stats := s.pageCache.Stats()
	return PageCacheStats{
		Entries:      stats.Entries,
		Max:          stats.Max,
		Bytes:        stats.Bytes,
		MaxBytes:     stats.MaxBytes,
		RAMEntries:   stats.RAMEntries,
		DiskEntries:  stats.DiskEntries,
		DiskBytes:    stats.DiskBytes,
		DiskMaxBytes: stats.DiskMaxBytes,
	}
}

func (s *BrowserService) ClearPageCache() int {
	if s == nil || s.pageCache == nil {
		return 0
	}
	cleared := s.pageCache.Clear()
	s.log("info", "page cache cleared", "")
	return cleared
}

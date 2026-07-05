// SPDX-License-Identifier: MIT
package app

type PageCacheStats struct {
	Entries int `json:"entries"`
	Max     int `json:"max"`
}

func (s *BrowserService) GetPageCacheStats() PageCacheStats {
	if s == nil || s.pageCache == nil {
		return PageCacheStats{Max: 128}
	}
	return PageCacheStats{
		Entries: s.pageCache.Len(),
		Max:     s.pageCache.Max(),
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

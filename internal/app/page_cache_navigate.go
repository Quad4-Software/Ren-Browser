// SPDX-License-Identifier: MIT
package app

import (
	"context"

	"renbrowser/internal/cache"
	"renbrowser/internal/content"
	"renbrowser/internal/limits"
	"renbrowser/internal/nomadnet"
)

func (s *BrowserService) pageResponseFromCache(url string, parsed nomadnet.PageURL, entry cache.Entry) PageResponse {
	rendered := content.Render(parsed.Path, entry.Body, parsed.NodeHash)
	resp := PageResponse{
		URL:         url,
		NodeHash:    parsed.NodeHash,
		Path:        parsed.Path,
		ContentType: rendered.Kind,
		HTML:        rendered.HTML,
		Raw:         rendered.Raw,
		PageFG:      rendered.PageFG,
		PageBG:      rendered.PageBG,
		FromCache:   true,
		CachedAt:    entry.StoredAt.UnixMilli(),
		Hops:        s.hopsForNode(parsed.NodeHash),
	}
	s.setLastPage(resp)
	s.recordNetwork(resp)
	s.log("info", "page cache hit", url)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
	}
	return resp
}

func (s *BrowserService) tryRefreshCachedPage(
	ctx context.Context,
	url string,
	parsed nomadnet.PageURL,
	fallback cache.Entry,
) *PageResponse {
	fetchPath, fetchReq, cancelled := s.runFetchHooks(parsed.NodeHash, parsed.Path, parsed.Request)
	if cancelled {
		return nil
	}
	fetch := s.fetchWithRetry(ctx, parsed.NodeHash, fetchPath, fetchReq)
	if fetch.Error != "" || len(fetch.Body) == 0 {
		s.log("info", "stale cache refresh failed, using cached copy", fetch.Error)
		return nil
	}
	rendered := content.Render(fetch.Path, fetch.Body, fetch.NodeHash)
	resp := PageResponse{
		URL:         url,
		NodeHash:    fetch.NodeHash,
		Path:        fetch.Path,
		ContentType: rendered.Kind,
		HTML:        rendered.HTML,
		Raw:         rendered.Raw,
		PageFG:      rendered.PageFG,
		PageBG:      rendered.PageBG,
		DurationMs:  fetch.DurationMs,
		Hops:        fetch.Hops,
		Interface:   fetch.Interface,
	}
	if s.GetBrowserPrefs().PageCacheEnabled && len(fetch.Body) <= limits.MaxPageBytes() {
		s.pageCache.Put(fetch.NodeHash, fetch.Path, parsed.Request, fetch.Body, rendered.Kind)
	}
	s.log("info", "stale cache refreshed", url)
	s.setLastPage(resp)
	s.recordNetwork(resp)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
	}
	_ = fallback
	return &resp
}

// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"renbrowser/internal/limits"
	"renbrowser/internal/nomadnet"
)

// NodeImageResult carries a fetched node image back to the frontend as
// base64 so it can be rendered from an in-memory blob URL.
type NodeImageResult struct {
	Data  string `json:"data"`
	Mime  string `json:"mime"`
	Bytes int    `json:"bytes"`
	Name  string `json:"name,omitempty"`
}

var (
	nodeImageHashRe  = regexp.MustCompile(`(?i)^[a-f0-9]{32}$`)
	nodeImageMediaRe = regexp.MustCompile(`(?i)\.(webp|png|jpe?g|bmp|gif|tiff)$`)
)

// validateNodeImagePath mirrors the micron-parser-go image rules: only
// /media/ paths with a raster extension and /file/ paths ending in .webp
// are fetchable as inline images.
func validateNodeImagePath(path string) error {
	if path == "" {
		return errors.New("empty image path")
	}
	p := path
	if i := strings.IndexByte(p, '`'); i >= 0 {
		p = p[:i]
	}
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	if i := strings.IndexByte(p, '#'); i >= 0 {
		p = p[:i]
	}
	p = strings.TrimSpace(p)
	if strings.HasPrefix(p, "/") {
		p = p[1:]
	}
	if strings.HasPrefix(p, "media/") {
		if !nodeImageMediaRe.MatchString(p) {
			return fmt.Errorf("unsupported image type for %q", path)
		}
	} else if strings.HasPrefix(p, "file/") {
		if !strings.HasSuffix(strings.ToLower(p), ".webp") {
			return fmt.Errorf("/file/ images must be .webp, got %q", path)
		}
	} else {
		return fmt.Errorf("image path must be under /media/ or /file/, got %q", path)
	}
	if strings.Contains(p, "..") {
		return errors.New("image path contains traversal")
	}
	for _, r := range p {
		if r < 32 || strings.ContainsRune(`<>"|?*`, r) {
			return fmt.Errorf("image path contains unsafe characters: %q", path)
		}
	}
	return nil
}

// sniffNodeImageMime rejects payloads that are not raster images. SVG is
// deliberately excluded: it can carry scriptable content.
func sniffNodeImageMime(body []byte) (string, error) {
	if len(body) == 0 {
		return "", errors.New("empty image response")
	}
	mime := http.DetectContentType(body)
	switch mime {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp", "image/tiff":
		return mime, nil
	}
	return "", fmt.Errorf("response is not a raster image (content-type %q)", mime)
}

// nodeImageCacheReq keys cached images by the page-declared key and profile
// so nodes that vary output per key get distinct entries.
func nodeImageCacheReq(imageKey, profile string) nomadnet.RequestData {
	if imageKey == "" && profile == "" {
		return nomadnet.RequestData{}
	}
	vars := make(map[string]string, 2)
	if imageKey != "" {
		vars["k"] = imageKey
	}
	if profile != "" {
		vars["profile"] = profile
	}
	return nomadnet.RequestData{Vars: vars}
}

// FetchNodeImage downloads a /media/ or /file/*.webp image from a mesh node
// for opt-in inline display. The response is capped at limits.MaxAssetBytes
// during receipt and must sniff as a raster image. imageKey and profile are
// the page-declared hints forwarded to the node's /media handler. Reload
// bypasses the local image cache and refreshes the stored entry.
func (s *BrowserService) FetchNodeImage(rawURL string, imageKey string, profile string, reload bool) (NodeImageResult, error) {
	parsed, err := nomadnet.ParseURL(rawURL)
	if err != nil {
		return NodeImageResult{}, err
	}
	if !nodeImageHashRe.MatchString(parsed.NodeHash) {
		return NodeImageResult{}, fmt.Errorf("invalid node hash in %q", rawURL)
	}
	if err := validateNodeImagePath(parsed.Path); err != nil {
		return NodeImageResult{}, err
	}

	cacheReq := nodeImageCacheReq(imageKey, profile)
	cacheEnabled := s.imageCache != nil && s.GetBrowserPrefs().PageCacheEnabled
	if cacheEnabled && !reload {
		if entry, ok := s.imageCache.Get(parsed.NodeHash, parsed.Path, cacheReq); ok && len(entry.Body) > 0 {
			return NodeImageResult{
				Data:  base64.StdEncoding.EncodeToString(entry.Body),
				Mime:  entry.ContentType,
				Bytes: len(entry.Body),
			}, nil
		}
	}

	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return NodeImageResult{}, errors.New("reticulum not ready")
	}

	maxBytes := limits.MaxAssetBytes()
	ctx, cancel := context.WithTimeout(context.Background(), s.fetchBudget(parsed.NodeHash, parsed.Path))
	defer cancel()

	hooks := s.fileFetchHooks(rawURL)
	if s.app != nil {
		hooks = mergeFetchHooks(hooks, &nomadnet.FetchHooks{
			OnProgress: func(p nomadnet.FetchProgress) {
				s.app.Event.Emit("micron:image-progress", map[string]any{
					"url":      rawURL,
					"received": p.Received,
					"total":    p.Total,
				})
			},
		})
	}

	var fetch nomadnet.FetchResult
	if strings.HasPrefix(parsed.Path, "/media/") {
		fetch = stack.Browser().FetchMedia(ctx, parsed.NodeHash, parsed.Path, imageKey, profile, maxBytes, hooks)
	} else {
		fetch = stack.Browser().FetchLimited(
			ctx, parsed.NodeHash, parsed.Path, nomadnet.RequestData{}, maxBytes, hooks,
		)
	}
	if fetch.Error != "" {
		s.log("error", "image fetch failed", fmt.Sprintf("%s: %s", rawURL, fetch.Error))
		return NodeImageResult{}, errors.New(fetch.Error)
	}
	if len(fetch.Body) == 0 {
		return NodeImageResult{}, errors.New("empty image response")
	}
	if len(fetch.Body) > maxBytes {
		return NodeImageResult{}, fmt.Errorf("image too large: %d bytes (limit %d)", len(fetch.Body), maxBytes)
	}
	mime, err := sniffNodeImageMime(fetch.Body)
	if err != nil {
		return NodeImageResult{}, err
	}
	if cacheEnabled {
		s.imageCache.Put(parsed.NodeHash, parsed.Path, cacheReq, fetch.Body, mime)
	}
	return NodeImageResult{
		Data:  base64.StdEncoding.EncodeToString(fetch.Body),
		Mime:  mime,
		Bytes: len(fetch.Body),
		Name:  fetch.FileName,
	}, nil
}

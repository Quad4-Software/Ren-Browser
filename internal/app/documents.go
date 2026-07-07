// SPDX-License-Identifier: MIT
package app

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"renbrowser/internal/limits"
	"renbrowser/internal/nomadnet"
)

func isDocumentURL(raw string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(raw)), "document:")
}

func parseDocumentPath(rawURL string) (string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if !isDocumentURL(trimmed) {
		return "", fmt.Errorf("not a document URL")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(parsed.Query().Get("path"))
	if path == "" {
		return "", fmt.Errorf("document path is required")
	}
	return path, nil
}

func documentURL(path string) string {
	return "document:?path=" + url.QueryEscape(path)
}

func (s *BrowserService) documentPage(rawURL string, pushHistory bool) PageResponse {
	path, err := parseDocumentPath(rawURL)
	resp := PageResponse{URL: rawURL, Path: path}
	if err != nil {
		applyPageError(&resp, err.Error(), nil)
		s.setLastPage(resp)
		return resp
	}
	body, err := s.readValidatedDownloadFile(path)
	if err != nil {
		applyPageError(&resp, err.Error(), nil)
		s.setLastPage(resp)
		return resp
	}
	if !applyDocumentBody(&resp, path, body) {
		applyPageError(&resp, "unsupported document format", nil)
		s.setLastPage(resp)
		return resp
	}
	if pushHistory {
		s.pushHistory(rawURL)
		title := filepath.Base(path)
		_ = s.store.AddHistory(rawURL, title, "")
	}
	s.log("info", "document opened", path)
	s.setLastPage(resp)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
	}
	return resp
}

func (s *BrowserService) readValidatedDownloadFile(path string) ([]byte, error) {
	if err := s.validateDownloadPath(path); err != nil {
		return nil, fmt.Errorf("cannot read document outside download folder: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("document not found: %s", path)
		}
		return nil, err
	}
	max := limits.MaxDocumentViewBytes()
	if max <= 0 {
		max = limits.DefaultMaxPageBytes
	}
	if max > 0 && info.Size() > int64(max) {
		return nil, fmt.Errorf("document too large to view in browser (max %d bytes)", max)
	}
	return readFileLimited(path, max)
}

func applyDocumentBody(resp *PageResponse, path string, body []byte) bool {
	kind := nomadnet.DetectContentType(path, body)
	if !nomadnet.IsDocumentKind(kind) {
		return false
	}
	max := limits.MaxDocumentViewBytes()
	if max > 0 && len(body) > max {
		applyPageError(resp, fmt.Sprintf("document too large to view in browser (max %d bytes)", max), nil)
		resp.ContentType = kind
		return true
	}
	resp.ContentType = kind
	resp.BinaryB64 = base64.StdEncoding.EncodeToString(body)
	return true
}

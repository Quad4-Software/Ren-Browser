// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// DownloadRetryResult reports whether a user-initiated download retry started.
type DownloadRetryResult struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// RetryDownload restarts a failed, interrupted, or canceled download from the
// beginning. The source URL must still be known; the filename is recovered from
// the download entry, a stale path, or the URL when needed.
func (s *BrowserService) RetryDownload(id string) DownloadRetryResult {
	item, ok := s.downloads.findByID(id)
	if !ok {
		return DownloadRetryResult{Error: "download not found"}
	}
	switch item.Status {
	case DownloadStatusFailed, DownloadStatusInterrupted, DownloadStatusCanceled:
	default:
		return DownloadRetryResult{Error: "download cannot be retried in its current state"}
	}

	url, name, err := s.resolveRetryDownload(item)
	if err != nil {
		return DownloadRetryResult{Error: err.Error()}
	}

	s.downloads.dismiss(id)
	s.removePendingDownloadJob(url)
	s.startBackgroundDownload(url, name)
	return DownloadRetryResult{OK: true}
}

func (s *BrowserService) resolveRetryDownload(item ActiveDownload) (url, name string, err error) {
	url = strings.TrimSpace(item.URL)
	if url == "" {
		return "", "", errors.New("download source URL is missing")
	}
	name = s.resolveDownloadFilename(item)
	if name == "" {
		return "", "", errors.New("download filename could not be determined")
	}
	return url, name, nil
}

func (s *BrowserService) resolveDownloadFilename(item ActiveDownload) string {
	if name := strings.TrimSpace(item.Name); name != "" {
		return sanitizeDownloadFilename(name)
	}

	path := strings.TrimSpace(item.Path)
	if path != "" {
		if base := filepath.Base(path); base != "" && base != "." {
			return sanitizeDownloadFilename(base)
		}
	}

	if url := strings.TrimSpace(item.URL); url != "" {
		hint := downloadNameFromURL(url)
		if found := s.findExistingDownloadPath(hint); found != "" {
			return sanitizeDownloadFilename(filepath.Base(found))
		}
		return hint
	}

	return "download.bin"
}

func (s *BrowserService) findExistingDownloadPath(name string) string {
	name = sanitizeDownloadFilename(name)
	if name == "" {
		return ""
	}

	candidate := filepath.Join(s.GetDownloadDir(), name)
	if fileExists(candidate) {
		return candidate
	}

	for _, item := range s.loadDownloadHistory() {
		if !strings.EqualFold(filepath.Base(item.Path), name) {
			continue
		}
		path := filepath.Clean(item.Path)
		if fileExists(path) {
			return path
		}
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info != nil && !info.IsDir()
}

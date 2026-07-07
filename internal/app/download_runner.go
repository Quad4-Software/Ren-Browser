// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	downloadMaxAttempts = 4
	downloadRetryBase   = 2 * time.Second
)

func retriableDownloadError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{
		"timeout",
		"timed out",
		"empty response",
		"reticulum not ready",
		"link",
		"connection",
		"temporary",
		"node response",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func (s *BrowserService) runTrackedDownload(id, rawURL, name string) (string, error) {
	tracker := &downloadTracker{mgr: s.downloads, id: id}
	s.syncDownloadBackground()
	defer s.syncDownloadBackground()

	var lastErr error
	for attempt := 1; attempt <= downloadMaxAttempts; attempt++ {
		s.downloads.setAttempt(id, attempt)
		if attempt > 1 {
			s.downloads.setStatus(id, DownloadStatusRetrying)
			wait := downloadRetryBase * time.Duration(attempt-1)
			time.Sleep(wait)
		}

		fetch, err := s.fetchFileTracked(rawURL, tracker)
		if err == nil {
			if fetch.FileName != "" {
				name = sanitizeDownloadFilename(fetch.FileName)
				s.downloads.setName(id, name)
			}
			dest, writeErr := writeDownloadBytes(s.GetDownloadDir(), name, fetch.Body)
			if writeErr == nil {
				s.downloads.complete(id, dest, int64(len(fetch.Body)))
				s.recordDownload(dest)
				return dest, nil
			}
			lastErr = writeErr
		} else {
			lastErr = err
		}

		if !retriableDownloadError(lastErr) || attempt == downloadMaxAttempts {
			break
		}
	}

	s.downloads.fail(id, lastErr.Error())
	return "", lastErr
}

func (s *BrowserService) startBackgroundDownload(rawURL, name string) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return
	}
	if name == "" {
		name = downloadNameFromURL(rawURL)
	}
	id := s.downloads.start(rawURL, name)
	s.addPendingDownloadJob(rawURL, name)
	go func() {
		_, _ = s.runTrackedDownload(id, rawURL, name)
		s.removePendingDownloadJob(rawURL)
	}()
}

func (s *BrowserService) executeDownload(rawURL string) (string, error) {
	name := downloadNameFromURL(rawURL)
	id := s.downloads.start(rawURL, name)
	s.addPendingDownloadJob(rawURL, name)
	defer s.removePendingDownloadJob(rawURL)
	return s.runTrackedDownload(id, rawURL, name)
}

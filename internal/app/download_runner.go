// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrDownloadCanceled = errors.New("download canceled")

const (
	downloadMaxAttempts = 4
	downloadRetryBase   = 2 * time.Second
	// Cap parallel mesh file transfers so N large /file/ fetches cannot
	// each hold a full response buffer at once.
	maxConcurrentDownloads = 4
	// Legacy DownloadFile base64 path. UI uses DownloadToDir.
	maxDownloadFileB64Bytes = 16 * 1024 * 1024
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

	if err := s.acquireDownloadSlot(id); err != nil {
		return "", trackedDownloadFailure(s.downloads, id, err)
	}
	defer s.releaseDownloadSlot()

	var lastErr error
	for attempt := 1; attempt <= downloadMaxAttempts; attempt++ {
		if item, ok := s.downloads.findByID(id); ok && item.Status == DownloadStatusCanceled {
			return "", trackedDownloadFailure(s.downloads, id, ErrDownloadCanceled)
		}
		s.downloads.setAttempt(id, attempt)
		if attempt > 1 {
			s.downloads.setStatus(id, DownloadStatusRetrying)
			wait := downloadRetryBase * time.Duration(attempt-1)
			time.Sleep(wait)
			if item, ok := s.downloads.findByID(id); ok && item.Status == DownloadStatusCanceled {
				return "", trackedDownloadFailure(s.downloads, id, ErrDownloadCanceled)
			}
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

	return "", trackedDownloadFailure(s.downloads, id, lastErr)
}

func (s *BrowserService) acquireDownloadSlot(id string) error {
	if s == nil {
		return nil
	}
	slots := s.downloadSlots
	if slots == nil {
		return nil
	}
	for {
		if item, ok := s.downloads.findByID(id); ok && item.Status == DownloadStatusCanceled {
			return ErrDownloadCanceled
		}
		select {
		case slots <- struct{}{}:
			return nil
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func (s *BrowserService) releaseDownloadSlot() {
	if s == nil || s.downloadSlots == nil {
		return
	}
	select {
	case <-s.downloadSlots:
	default:
	}
}

func trackedDownloadFailure(m *downloadManager, id string, lastErr error) error {
	m.markFailedUnlessCanceled(id, lastErr)
	if lastErr == nil {
		return nil
	}
	if errors.Is(lastErr, context.Canceled) {
		if item, ok := m.findByID(id); ok && item.Status == DownloadStatusCanceled {
			return ErrDownloadCanceled
		}
	}
	return lastErr
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

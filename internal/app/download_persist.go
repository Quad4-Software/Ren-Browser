// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	downloadRecoveryKey     = "downloadRecovery"
	downloadInterruptedText = "Download interrupted when the app closed"
)

type downloadRecoveryRecord struct {
	ID          string         `json:"id"`
	URL         string         `json:"url"`
	Name        string         `json:"name"`
	Received    int64          `json:"received"`
	Total       int64          `json:"total"`
	Status      DownloadStatus `json:"status"`
	Error       string         `json:"error,omitempty"`
	Attempt     int            `json:"attempt,omitempty"`
	MaxAttempts int            `json:"maxAttempts,omitempty"`
	StartedAt   int64          `json:"startedAt"`
	UpdatedAt   int64          `json:"updatedAt"`
}

func recoveryStatus(status DownloadStatus) DownloadStatus {
	switch status {
	case DownloadStatusPending, DownloadStatusDownloading, DownloadStatusRetrying:
		return DownloadStatusInterrupted
	case DownloadStatusInterrupted, DownloadStatusFailed:
		return status
	default:
		return ""
	}
}

func (s *BrowserService) persistDownloadRecovery(items []ActiveDownload) {
	if s.store == nil {
		return
	}
	records := make([]downloadRecoveryRecord, 0)
	for _, item := range items {
		status := recoveryStatus(item.Status)
		if status == "" {
			continue
		}
		rec := activeDownloadToRecovery(item)
		rec.Status = status
		if status == DownloadStatusInterrupted && strings.TrimSpace(rec.Error) == "" {
			rec.Error = downloadInterruptedText
		}
		records = append(records, rec)
	}
	if len(records) == 0 {
		_ = s.store.SetSetting(downloadRecoveryKey, "")
		return
	}
	raw, err := json.Marshal(records)
	if err != nil {
		return
	}
	_ = s.store.SetSetting(downloadRecoveryKey, string(raw))
}

func (s *BrowserService) loadDownloadRecovery() []downloadRecoveryRecord {
	if s.store == nil {
		return nil
	}
	raw, err := s.store.GetSetting(downloadRecoveryKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil
	}
	var records []downloadRecoveryRecord
	if err := json.Unmarshal([]byte(raw), &records); err != nil {
		return nil
	}
	out := make([]downloadRecoveryRecord, 0, len(records))
	for _, rec := range records {
		if strings.TrimSpace(rec.URL) == "" {
			continue
		}
		status := recoveryStatus(rec.Status)
		if status == "" {
			continue
		}
		rec.Status = status
		if rec.Status == DownloadStatusInterrupted && strings.TrimSpace(rec.Error) == "" {
			rec.Error = downloadInterruptedText
		}
		if rec.MaxAttempts <= 0 {
			rec.MaxAttempts = downloadMaxAttempts
		}
		out = append(out, rec)
	}
	return out
}

func activeDownloadToRecovery(item ActiveDownload) downloadRecoveryRecord {
	return downloadRecoveryRecord{
		ID:          item.ID,
		URL:         item.URL,
		Name:        item.Name,
		Received:    item.Received,
		Total:       item.Total,
		Status:      item.Status,
		Error:       item.Error,
		Attempt:     item.Attempt,
		MaxAttempts: item.MaxAttempts,
		StartedAt:   item.StartedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func (s *BrowserService) reconcileDownloadsOnStartup() {
	records := s.loadDownloadRecovery()
	seen := make(map[string]struct{}, len(records))
	for _, rec := range records {
		seen[rec.URL] = struct{}{}
	}

	for _, job := range s.loadPendingDownloadJobs() {
		if _, ok := seen[job.URL]; ok {
			continue
		}
		now := time.Now().UnixMilli()
		records = append(records, downloadRecoveryRecord{
			ID:          fmt.Sprintf("dl-%d", s.downloads.nextRecoveryID()),
			URL:         job.URL,
			Name:        job.Name,
			Status:      DownloadStatusInterrupted,
			Error:       downloadInterruptedText,
			MaxAttempts: downloadMaxAttempts,
			StartedAt:   now,
			UpdatedAt:   now,
		})
	}
	s.savePendingDownloadJobs(nil)

	for _, rec := range records {
		s.downloads.importRecovery(rec)
	}
	s.persistDownloadRecovery(s.downloads.list())
	if len(records) > 0 {
		s.downloads.notify()
	}
}

func (m *downloadManager) nextRecoveryID() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	return m.nextID
}

func (m *downloadManager) importRecovery(rec downloadRecoveryRecord) {
	id := strings.TrimSpace(rec.ID)
	if id == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.items[id]; exists {
		return
	}
	if n := parseDownloadIDNum(id); n >= m.nextID {
		m.nextID = n
	}
	now := time.Now().UnixMilli()
	m.items[id] = &ActiveDownload{
		ID:          id,
		URL:         rec.URL,
		Name:        rec.Name,
		Received:    rec.Received,
		Total:       rec.Total,
		Status:      rec.Status,
		Error:       rec.Error,
		Attempt:     rec.Attempt,
		MaxAttempts: rec.MaxAttempts,
		StartedAt:   rec.StartedAt,
		UpdatedAt:   rec.UpdatedAt,
	}
	if m.items[id].StartedAt == 0 {
		m.items[id].StartedAt = now
	}
	if m.items[id].UpdatedAt == 0 {
		m.items[id].UpdatedAt = now
	}
	if m.items[id].MaxAttempts <= 0 {
		m.items[id].MaxAttempts = downloadMaxAttempts
	}
	m.order = append([]string{id}, m.order...)
}

func parseDownloadIDNum(id string) int64 {
	if !strings.HasPrefix(id, "dl-") {
		return 0
	}
	n, err := strconv.ParseInt(strings.TrimPrefix(id, "dl-"), 10, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func (m *downloadManager) shutdownInFlight(errMsg string) {
	m.mu.Lock()
	cancels := make([]context.CancelFunc, 0)
	for _, d := range m.items {
		switch d.Status {
		case DownloadStatusPending, DownloadStatusDownloading, DownloadStatusRetrying:
			d.Status = DownloadStatusInterrupted
			d.Error = errMsg
			d.UpdatedAt = time.Now().UnixMilli()
			if d.cancel != nil {
				cancels = append(cancels, d.cancel)
			}
		}
	}
	m.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
	if len(cancels) > 0 {
		m.notify()
	}
}

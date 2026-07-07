// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"strings"
)

const pendingDownloadsKey = "pendingDownloads"

type pendingDownloadJob struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

func (s *BrowserService) addPendingDownloadJob(url, name string) {
	url = strings.TrimSpace(url)
	if url == "" {
		return
	}
	jobs := s.loadPendingDownloadJobs()
	filtered := make([]pendingDownloadJob, 0, len(jobs)+1)
	for _, job := range jobs {
		if job.URL != url {
			filtered = append(filtered, job)
		}
	}
	filtered = append(filtered, pendingDownloadJob{URL: url, Name: name})
	s.savePendingDownloadJobs(filtered)
}

func (s *BrowserService) removePendingDownloadJob(url string) {
	url = strings.TrimSpace(url)
	if url == "" {
		return
	}
	jobs := s.loadPendingDownloadJobs()
	filtered := jobs[:0]
	for _, job := range jobs {
		if job.URL != url {
			filtered = append(filtered, job)
		}
	}
	if len(filtered) == len(jobs) {
		return
	}
	s.savePendingDownloadJobs(filtered)
}

func (s *BrowserService) loadPendingDownloadJobs() []pendingDownloadJob {
	if s.store == nil {
		return nil
	}
	raw, err := s.store.GetSetting(pendingDownloadsKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil
	}
	var jobs []pendingDownloadJob
	if err := json.Unmarshal([]byte(raw), &jobs); err != nil {
		return nil
	}
	out := make([]pendingDownloadJob, 0, len(jobs))
	for _, job := range jobs {
		if strings.TrimSpace(job.URL) == "" {
			continue
		}
		out = append(out, job)
	}
	return out
}

func (s *BrowserService) savePendingDownloadJobs(jobs []pendingDownloadJob) {
	if s.store == nil {
		return
	}
	if len(jobs) == 0 {
		_ = s.store.SetSetting(pendingDownloadsKey, "")
		return
	}
	raw, err := json.Marshal(jobs)
	if err != nil {
		return
	}
	_ = s.store.SetSetting(pendingDownloadsKey, string(raw))
}

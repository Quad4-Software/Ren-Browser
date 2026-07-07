//go:build android

// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func (s *BrowserService) syncDownloadBackground() {
	if s.downloads.runningCount() > 0 {
		s.ensureDownloadForegroundService()
		return
	}
	s.stopDownloadForegroundService()
}

var downloadForegroundMu sync.Mutex
var downloadForegroundRunning bool

func (s *BrowserService) ensureDownloadForegroundService() {
	downloadForegroundMu.Lock()
	defer downloadForegroundMu.Unlock()
	if downloadForegroundRunning {
		return
	}
	payload, err := json.Marshal(map[string]string{
		"title": "Ren Browser",
		"text":  "Downloading over the mesh",
	})
	if err != nil {
		return
	}
	application.Android.StartForegroundService(string(payload))
	downloadForegroundRunning = true
}

func (s *BrowserService) stopDownloadForegroundService() {
	downloadForegroundMu.Lock()
	defer downloadForegroundMu.Unlock()
	if !downloadForegroundRunning {
		return
	}
	application.Android.StopForegroundService()
	downloadForegroundRunning = false
}

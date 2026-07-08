// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"quad4/reticulum-go/pkg/reticulumconfig"
	"renbrowser/internal/rns"
)

type CheckStatus struct {
	Passed bool   `json:"passed"`
	Reason string `json:"reason,omitempty"`
}

type SelfCheckResult struct {
	StackUp       CheckStatus `json:"stackUp"`
	ConfigGood    CheckStatus `json:"configGood"`
	DBGood        CheckStatus `json:"dbGood"`
	ReadWriteGood CheckStatus `json:"readWriteGood"`
	DownloadsGood CheckStatus `json:"downloadsGood"`
	AllPassed     bool        `json:"allPassed"`
}

// RunSelfCheck performs a comprehensive internal health check of the application components.
func (s *BrowserService) RunSelfCheck() SelfCheckResult {
	res := SelfCheckResult{
		AllPassed: true,
	}

	// 1. Stack up check
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()

	if stack == nil {
		res.StackUp = CheckStatus{Passed: false, Reason: "Reticulum stack is not initialized"}
		res.AllPassed = false
	} else if !stack.IsStarted() {
		res.StackUp = CheckStatus{Passed: false, Reason: "Reticulum stack is initialized but not started"}
		res.AllPassed = false
	} else {
		res.StackUp = CheckStatus{Passed: true}
	}

	// 2. Config good check
	configPath := ""
	if stack != nil {
		configPath = stack.ConfigPath()
	}
	if configPath == "" {
		configPath = filepath.Join(rns.DefaultConfigDir(), "config")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		res.ConfigGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Reticulum config file not found at: %s", configPath)}
		res.AllPassed = false
	} else {
		_, err := reticulumconfig.LoadConfig(configPath)
		if err != nil {
			res.ConfigGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Failed to parse Reticulum config: %v", err)}
			res.AllPassed = false
		} else {
			res.ConfigGood = CheckStatus{Passed: true}
		}
	}

	// 3. DB good check
	health := s.GetStoreHealth()
	if !health.OK {
		res.DBGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Database health check failed: %s (%s)", health.Detail, health.Kind)}
		res.AllPassed = false
	} else {
		res.DBGood = CheckStatus{Passed: true}
	}

	// 4. Read/Write good check
	if s.store == nil {
		res.ReadWriteGood = CheckStatus{Passed: false, Reason: "Database store is unavailable"}
		res.AllPassed = false
	} else {
		testKey := "self_check_last_run"
		testVal := time.Now().UTC().Format(time.RFC3339)
		if err := s.store.SetSetting(testKey, testVal); err != nil {
			res.ReadWriteGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Database write failed: %v", err)}
			res.AllPassed = false
		} else {
			val, err := s.store.GetSetting(testKey)
			if err != nil {
				res.ReadWriteGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Database read failed: %v", err)}
				res.AllPassed = false
			} else if val != testVal {
				res.ReadWriteGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Database read value mismatch: expected %q, got %q", testVal, val)}
				res.AllPassed = false
			} else {
				res.ReadWriteGood = CheckStatus{Passed: true}
			}
		}
	}

	// 5. Check folder read/write for downloads
	downloadDir := s.GetDownloadDir()
	if downloadDir == "" {
		res.DownloadsGood = CheckStatus{Passed: false, Reason: "Downloads directory path is empty"}
		res.AllPassed = false
	} else {
		if err := os.MkdirAll(downloadDir, 0o755); err != nil {
			res.DownloadsGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Failed to create downloads directory: %v", err)}
			res.AllPassed = false
		} else {
			tempFile := filepath.Join(downloadDir, ".self_check_temp")
			testData := []byte("self_check_data")
			if err := os.WriteFile(tempFile, testData, 0o644); err != nil {
				res.DownloadsGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Failed to write to downloads directory: %v", err)}
				res.AllPassed = false
			} else {
				readData, err := os.ReadFile(tempFile)
				_ = os.Remove(tempFile) // Clean up immediately

				if err != nil {
					res.DownloadsGood = CheckStatus{Passed: false, Reason: fmt.Sprintf("Failed to read from downloads directory: %v", err)}
					res.AllPassed = false
				} else if string(readData) != string(testData) {
					res.DownloadsGood = CheckStatus{Passed: false, Reason: "Downloads read/write data mismatch"}
					res.AllPassed = false
				} else {
					res.DownloadsGood = CheckStatus{Passed: true}
				}
			}
		}
	}

	return res
}

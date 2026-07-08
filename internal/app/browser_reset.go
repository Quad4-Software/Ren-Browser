// SPDX-License-Identifier: MIT
package app

import (
	"log"
	"os"
	"path/filepath"

	"renbrowser/internal/brand"
	"renbrowser/internal/paths"
	"renbrowser/internal/rns"
)

// ResetBrowser performs a full reset of the browser data and configuration.
// It closes the database, stops Reticulum, deletes the data directory and
// Reticulum config, and then exits the application.
func (s *BrowserService) ResetBrowser() {
	s.mu.Lock()
	if s.shuttingDown {
		s.mu.Unlock()
		return
	}
	s.shuttingDown = true
	s.mu.Unlock()

	log.Println("ResetBrowser: starting full reset...")

	// 1. Close store
	if s.store != nil {
		log.Println("ResetBrowser: closing store...")
		_ = s.store.Close()
	}

	// 2. Stop Reticulum
	if s.stack != nil {
		log.Println("ResetBrowser: stopping Reticulum...")
		_ = s.stack.Stop()
	}

	// 3. Delete data directory
	dataDir := filepath.Join(paths.DataRoot(), brand.DataDirName)
	log.Printf("ResetBrowser: removing browser data directory: %s", dataDir)
	if err := os.RemoveAll(dataDir); err != nil {
		log.Printf("ResetBrowser warning: failed to remove browser data directory: %v", err)
	}

	// 4. Delete Reticulum config
	configPath := s.ConfigPath()
	if configPath != "" {
		log.Printf("ResetBrowser: removing Reticulum config: %s", configPath)
		if err := os.RemoveAll(filepath.Dir(configPath)); err != nil {
			log.Printf("ResetBrowser warning: failed to remove Reticulum config directory: %v", err)
		}
	} else {
		retConfigDir := rns.DefaultConfigDir()
		log.Printf("ResetBrowser: removing default Reticulum config directory: %s", retConfigDir)
		_ = os.RemoveAll(retConfigDir)
	}

	log.Println("ResetBrowser: reset complete, exiting app...")

	// 5. Quit app
	if s.app != nil {
		s.app.Quit()
	} else {
		os.Exit(0)
	}
}

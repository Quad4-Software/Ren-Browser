// SPDX-License-Identifier: MIT
package bootstrap

import (
	"log"
	"os"
	"path/filepath"

	"renbrowser/internal/brand"
	"renbrowser/internal/config"
	"renbrowser/internal/paths"
	"renbrowser/internal/rns"
)

// HandleResetIfNeeded checks if a browser reset has been requested via flags or
// environment variables, and if so, deletes the browser config and data
// directories to revert to clean defaults.
func HandleResetIfNeeded(cfg config.Runtime) {
	if !cfg.Reset {
		return
	}

	log.Println("Resetting browser config, cache, database, and settings...")

	dataDir := filepath.Join(paths.DataRoot(), brand.DataDirName)
	log.Printf("Removing browser data directory: %s", dataDir)
	if err := os.RemoveAll(dataDir); err != nil {
		log.Printf("Warning: failed to remove browser data directory: %v", err)
	}

	retConfigPath := cfg.ReticulumConfig
	if retConfigPath != "" {
		log.Printf("Removing custom Reticulum config file: %s", retConfigPath)
		if err := os.Remove(retConfigPath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to remove custom Reticulum config: %v", err)
		}
	} else {
		retConfigDir := rns.DefaultConfigDir()
		log.Printf("Removing Reticulum config directory: %s", retConfigDir)
		if err := os.RemoveAll(retConfigDir); err != nil {
			log.Printf("Warning: failed to remove Reticulum config directory: %v", err)
		}
	}
	log.Println("Browser reset complete.")
}

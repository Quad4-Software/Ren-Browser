// SPDX-License-Identifier: MIT
package rns

import (
	"os"
	"path/filepath"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/reticulumconfig"

	"renbrowser/internal/brand"
)

func ensureRenBrowserConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := reticulumconfig.CreateDefaultConfig(path); err != nil {
			return err
		}
		cfg, err := reticulumconfig.LoadConfig(path)
		if err != nil {
			return err
		}
		applyRenBrowserDefaults(cfg)
		// Explicitly set these only on first creation
		cfg.EnableTransport = false
		cfg.ShareInstance = false
		cfg.MaxInMemoryPaths = renBrowserDefaultMaxPaths
		cfg.MaxInMemoryKnownDestinations = renBrowserDefaultMaxKnownDests
		cfg.SoftMemoryLimitBytes = renBrowserDefaultSoftMemory
		cfg.MaxPacketHashlist = renBrowserDefaultHashlist
		return reticulumconfig.SaveConfig(cfg)
	}
	return nil
}

const (
	renBrowserDefaultMaxPaths      = 25_000
	renBrowserDefaultMaxKnownDests = 25_000
	renBrowserDefaultSoftMemory    = 192 << 20
	renBrowserDefaultHashlist      = 100_000
)

func applyRenBrowserDefaults(cfg *common.ReticulumConfig) {
	if cfg == nil {
		return
	}
	if cfg.AppName == "" || cfg.AppName == "Go Client" {
		cfg.AppName = brand.DisplayName
	}
	// Client mode keeps tighter RAM budgets even when the config file omits
	// the new keys. Explicit positive values in the file win.
	if !cfg.EnableTransport {
		if cfg.MaxInMemoryPaths == 0 {
			cfg.MaxInMemoryPaths = renBrowserDefaultMaxPaths
		}
		if cfg.MaxInMemoryKnownDestinations == 0 {
			cfg.MaxInMemoryKnownDestinations = renBrowserDefaultMaxKnownDests
		}
		if cfg.MaxPacketHashlist == 0 {
			cfg.MaxPacketHashlist = renBrowserDefaultHashlist
		}
		if cfg.SoftMemoryLimitBytes == 0 {
			cfg.SoftMemoryLimitBytes = renBrowserDefaultSoftMemory
		}
	}
}

func loadConfig(override string) (*common.ReticulumConfig, error) {
	path := override
	if path == "" {
		path = filepath.Join(DefaultConfigDir(), "config")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := ensureRenBrowserConfig(path); err != nil {
		return nil, err
	}
	cfg, err := reticulumconfig.LoadConfig(path)
	if err != nil {
		return nil, err
	}
	applyRenBrowserDefaults(cfg)
	cfg.ConfigPath = filepath.Clean(path)
	return cfg, nil
}

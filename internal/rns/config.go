package rns

import (
	"os"
	"path/filepath"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/reticulumconfig"
)

func ensureRenBrowserConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := reticulumconfig.CreateDefaultConfig(path); err != nil {
			return err
		}
		cfg, err := reticulumconfig.LoadConfig(path)
		if err != nil {
			return err
		}
		applyRenBrowserDefaults(cfg)
		return reticulumconfig.SaveConfig(cfg)
	}
	return nil
}

func applyRenBrowserDefaults(cfg *common.ReticulumConfig) {
	if cfg == nil {
		return
	}
	if cfg.AppName == "" || cfg.AppName == "Go Client" {
		cfg.AppName = "Ren Browser"
	}

	if _, exists := cfg.Interfaces["RNS Testnet TCP"]; !exists {
		cfg.Interfaces["RNS Testnet TCP"] = &common.InterfaceConfig{
			Name:        "RNS Testnet TCP",
			Type:        "TCPClientInterface",
			Enabled:     false,
			TargetHost:  "rns.michmesh.net",
			TargetPort:  7822,
			I2PTunneled: false,
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

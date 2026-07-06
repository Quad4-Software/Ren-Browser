// SPDX-License-Identifier: MIT
package rns

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/reticulumconfig"

	"renbrowser/internal/brand"
)

const defaultCommunityInterfaceCount = 4

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
		seedCommunityInterfaces(cfg)
		return reticulumconfig.SaveConfig(cfg)
	}
	return nil
}

func applyRenBrowserDefaults(cfg *common.ReticulumConfig) {
	if cfg == nil {
		return
	}
	if cfg.AppName == "" || cfg.AppName == "Go Client" {
		cfg.AppName = brand.DisplayName
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

func seedCommunityInterfaces(cfg *common.ReticulumConfig) {
	if testing.Testing() {
		return
	}
	if cfg == nil {
		return
	}
	items, err := FetchCommunityInterfaces(nil)
	if err != nil {
		return
	}
	tcp := FilterTCPClientInterfaces(items.Items)
	if len(tcp) == 0 {
		return
	}
	rand.Shuffle(len(tcp), func(i, j int) { tcp[i], tcp[j] = tcp[j], tcp[i] })
	if len(tcp) > defaultCommunityInterfaceCount {
		tcp = tcp[:defaultCommunityInterfaceCount]
	}
	for _, item := range tcp {
		if strings.TrimSpace(item.Config) == "" {
			continue
		}
		ifaces, err := parseInterfaceFragment(item.Config)
		if err != nil {
			continue
		}
		for name, iface := range ifaces {
			if iface == nil {
				continue
			}
			if _, exists := cfg.Interfaces[name]; exists {
				continue
			}
			iface.Name = name
			iface.Enabled = true
			cfg.Interfaces[name] = iface
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

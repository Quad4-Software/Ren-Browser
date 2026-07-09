// SPDX-License-Identifier: MIT
package rns

import (
	"strings"

	"quad4/reticulum-go/pkg/common"
)

// EffectiveInterfaceConfig returns the runtime interface config.
// With reticulum-go >= 0.9.9, BackboneClientInterface is native; no rewrite.
func EffectiveInterfaceConfig(cfg *common.InterfaceConfig) *common.InterfaceConfig {
	return cfg
}

func migrateInterfaceConfigs(cfg *common.ReticulumConfig) bool {
	return false
}

func IsBackboneInterface(item CommunityInterface) bool {
	if strings.TrimSpace(item.Config) == "" {
		return false
	}
	t := strings.ToLower(strings.TrimSpace(item.Type))
	tn := strings.ToLower(strings.TrimSpace(item.TypeName))
	if t == "backbone" || strings.Contains(t, "backbone") {
		return true
	}
	return strings.Contains(tn, "backbone")
}

func IsPipeInterface(item CommunityInterface) bool {
	if strings.TrimSpace(item.Config) == "" {
		return false
	}
	t := strings.ToLower(strings.TrimSpace(item.Type))
	tn := strings.ToLower(strings.TrimSpace(item.TypeName))
	if t == "pipe" || strings.Contains(t, "pipe") {
		return true
	}
	return strings.Contains(tn, "pipe")
}

func IsI2PInterface(item CommunityInterface) bool {
	if strings.TrimSpace(item.Config) == "" {
		return false
	}
	if strings.Contains(strings.ToLower(strings.TrimSpace(item.Network)), "i2p") {
		return true
	}
	t := strings.ToLower(strings.TrimSpace(item.Type))
	tn := strings.ToLower(strings.TrimSpace(item.TypeName))
	return strings.Contains(t, "i2p") || strings.Contains(tn, "i2p")
}

func isOverlayInterface(item CommunityInterface) bool {
	net := strings.ToLower(strings.TrimSpace(item.Network))
	if net == "yggdrasil" || net == "onion" || net == "tor" {
		return true
	}
	t := strings.ToLower(strings.TrimSpace(item.Type))
	if strings.Contains(t, "yggdrasil") || strings.Contains(t, "onion") || strings.Contains(t, "tor") {
		return true
	}
	tn := strings.ToLower(strings.TrimSpace(item.TypeName))
	if strings.Contains(tn, "yggdrasil") || strings.Contains(tn, "onion") || strings.Contains(tn, "tor") {
		return true
	}
	host := strings.ToLower(strings.TrimSpace(item.Host))
	if strings.HasSuffix(host, ".onion") {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(item.Name))
	if strings.Contains(name, "yggdrasil") || strings.Contains(name, "onion") || strings.Contains(name, "tor") {
		return true
	}
	return false
}

// FilterSeedableInterfaces keeps TCP client, backbone, pipe, and I2P entries
// that ship with a config snippet. Overlay networks other than I2P stay excluded.
func FilterSeedableInterfaces(items []CommunityInterface) []CommunityInterface {
	out := make([]CommunityInterface, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Config) == "" {
			continue
		}
		if isOverlayInterface(item) {
			continue
		}
		if IsTCPClientInterface(item) || IsBackboneInterface(item) || IsPipeInterface(item) || IsI2PInterface(item) {
			out = append(out, item)
		}
	}
	return out
}

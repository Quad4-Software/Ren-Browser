// SPDX-License-Identifier: MIT
package rns

import (
	"strings"

	"quad4/reticulum-go/pkg/common"
)

func isBackboneInterfaceType(t string) bool {
	lower := strings.ToLower(strings.TrimSpace(t))
	return lower == "backboneinterface" || strings.Contains(lower, "backbone")
}

func backboneToTCPClient(cfg *common.InterfaceConfig) (*common.InterfaceConfig, bool) {
	if cfg == nil || !isBackboneInterfaceType(cfg.Type) {
		return cfg, false
	}
	host := strings.TrimSpace(cfg.TargetHost)
	if host == "" {
		host = strings.TrimSpace(cfg.Address)
	}
	if host == "" {
		host = strings.TrimSpace(cfg.TargetAddress)
	}
	port := cfg.TargetPort
	if port == 0 {
		port = cfg.Port
	}
	if host == "" || port <= 0 {
		return cfg, false
	}
	adapted := *cfg
	adapted.Type = "TCPClientInterface"
	adapted.TargetHost = host
	adapted.TargetPort = port
	adapted.Address = ""
	adapted.Port = 0
	return &adapted, true
}

// EffectiveInterfaceConfig returns the runtime interface config for the current platform.
func EffectiveInterfaceConfig(cfg *common.InterfaceConfig) *common.InterfaceConfig {
	if cfg == nil || !usesBackboneTCPFallback() {
		return cfg
	}
	adapted, ok := backboneToTCPClient(cfg)
	if !ok {
		return cfg
	}
	return adapted
}

func migrateInterfaceConfigs(cfg *common.ReticulumConfig) bool {
	if cfg == nil || !usesBackboneTCPFallback() {
		return false
	}
	changed := false
	for name, iface := range cfg.Interfaces {
		adapted, ok := backboneToTCPClient(iface)
		if !ok {
			continue
		}
		cfg.Interfaces[name] = adapted
		changed = true
	}
	return changed
}

func rewriteBackboneSnippetToTCP(snippet string) string {
	lines := strings.Split(snippet, "\n")
	for i, line := range lines {
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "type") || !strings.Contains(lower, "backbone") {
			continue
		}
		replaced := strings.Replace(line, "BackboneInterface", "TCPClientInterface", 1)
		replaced = strings.Replace(replaced, "backboneinterface", "TCPClientInterface", 1)
		lines[i] = replaced
	}
	return strings.Join(lines, "\n")
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

func AdaptCommunityBackboneItem(item CommunityInterface) CommunityInterface {
	item.Type = "tcp"
	item.TypeName = "TCPClientInterface"
	item.Config = rewriteBackboneSnippetToTCP(normalizeConfigSnippet(item.Config))
	return item
}

func AdaptCommunityItemsForPlatform(items []CommunityInterface) []CommunityInterface {
	if !usesBackboneTCPFallback() || len(items) == 0 {
		return items
	}
	out := make([]CommunityInterface, len(items))
	for i, item := range items {
		if IsBackboneInterface(item) {
			out[i] = AdaptCommunityBackboneItem(item)
		} else {
			out[i] = item
		}
	}
	return out
}

func FilterSeedableInterfaces(items []CommunityInterface) []CommunityInterface {
	out := make([]CommunityInterface, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Config) == "" {
			continue
		}
		if IsTCPClientInterface(item) {
			out = append(out, item)
			continue
		}
		if usesBackboneTCPFallback() && IsBackboneInterface(item) {
			out = append(out, AdaptCommunityBackboneItem(item))
		}
	}
	return out
}

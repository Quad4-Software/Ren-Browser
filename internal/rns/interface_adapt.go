// SPDX-License-Identifier: MIT
package rns

import (
	"strings"

	"quad4/reticulum-go/pkg/common"
)

// FIXME(user1): Remove this entire workaround once vendored Reticulum-Go ships
// BackboneClientInterface and RenBrowser no longer needs to rewrite backbone configs.
// Until then, community backbone entries (remote/target_host) dial out as TCPClientInterface.

func usesBackboneTCPFallback() bool { return true }

func isBackboneInterfaceType(t string) bool {
	lower := strings.ToLower(strings.TrimSpace(t))
	return lower == "backboneinterface" || strings.Contains(lower, "backbone")
}

// FIXME(user1): Replace with BackboneClientInterface from Reticulum-Go when ready.
func backboneToTCPClient(cfg *common.InterfaceConfig) (*common.InterfaceConfig, bool) {
	if cfg == nil || !isBackboneInterfaceType(cfg.Type) {
		return cfg, false
	}
	host := strings.TrimSpace(cfg.TargetHost)
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
// FIXME(user1): Drop this shim when BackboneClientInterface is available in Reticulum-Go.
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

// FIXME(user1): Remove after BackboneClientInterface lands; migrates saved backbone client configs.
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

// FIXME(user1): Remove when backbone snippets can stay BackboneInterface in config files.
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

// FIXME(user1): Remove when community directory can expose BackboneClientInterface natively.
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

func isOverlayInterface(item CommunityInterface) bool {
	net := strings.ToLower(strings.TrimSpace(item.Network))
	if net == "yggdrasil" || net == "i2p" || net == "onion" || net == "tor" {
		return true
	}
	t := strings.ToLower(strings.TrimSpace(item.Type))
	if strings.Contains(t, "i2p") || strings.Contains(t, "yggdrasil") || strings.Contains(t, "onion") || strings.Contains(t, "tor") {
		return true
	}
	tn := strings.ToLower(strings.TrimSpace(item.TypeName))
	if strings.Contains(tn, "i2p") || strings.Contains(tn, "yggdrasil") || strings.Contains(tn, "onion") || strings.Contains(tn, "tor") {
		return true
	}
	host := strings.ToLower(strings.TrimSpace(item.Host))
	if strings.HasSuffix(host, ".i2p") || strings.HasSuffix(host, ".onion") {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(item.Name))
	if strings.Contains(name, "i2p") || strings.Contains(name, "yggdrasil") || strings.Contains(name, "onion") || strings.Contains(name, "tor") {
		return true
	}
	return false
}

func FilterSeedableInterfaces(items []CommunityInterface) []CommunityInterface {
	out := make([]CommunityInterface, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Config) == "" {
			continue
		}
		if isOverlayInterface(item) {
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

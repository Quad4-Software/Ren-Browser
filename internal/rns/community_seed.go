// SPDX-License-Identifier: MIT
package rns

import (
	"crypto/rand"
	"math/big"
	"net"
	"strconv"
	"strings"

	"quad4/reticulum-go/pkg/common"
)

// DefaultCommunityInterfaceCount is how many random community uplinks to seed
// on first-run / suggested setup. Six gives geographic diversity without
// opening too many long-lived TCP/backbone sockets.
const DefaultCommunityInterfaceCount = 6

func PickSeedableCommunityInterfaces(items []CommunityInterface, count int) []CommunityInterface {
	seedable := rankSeedableCommunityInterfaces(FilterSeedableInterfaces(items))
	if len(seedable) == 0 {
		return nil
	}
	cryptoShuffle(seedable)
	seedable = dedupeCommunityInterfaces(seedable)
	if count > 0 && len(seedable) > count {
		seedable = seedable[:count]
	}
	return seedable
}

func cryptoShuffle(items []CommunityInterface) {
	for i := len(items) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return
		}
		j := int(n.Int64())
		items[i], items[j] = items[j], items[i]
	}
}

// rankSeedableCommunityInterfaces prefers clearnet TCP/backbone that look
// healthy and skips obvious junk (offline, empty host, overlay networks).
func rankSeedableCommunityInterfaces(items []CommunityInterface) []CommunityInterface {
	out := make([]CommunityInterface, 0, len(items))
	for _, item := range items {
		if !isPreferredCommunitySeed(item) {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return items
	}
	return out
}

func isPreferredCommunitySeed(item CommunityInterface) bool {
	if strings.TrimSpace(item.Config) == "" {
		return false
	}
	status := strings.ToLower(strings.TrimSpace(item.Status))
	if status != "" && status != "online" {
		return false
	}
	netName := strings.ToLower(strings.TrimSpace(item.Network))
	if netName == "i2p" || netName == "yggdrasil" || netName == "onion" || netName == "tor" {
		return false
	}
	if IsI2PInterface(item) {
		return false
	}
	host := strings.TrimSpace(item.Host)
	if host == "" {
		host = hostFromConfigSnippet(item.Config)
	}
	if host == "" {
		return false
	}
	lowerHost := strings.ToLower(host)
	if strings.HasSuffix(lowerHost, ".i2p") || strings.HasSuffix(lowerHost, ".onion") {
		return false
	}
	return IsTCPClientInterface(item) || IsBackboneInterface(item)
}

func hostFromConfigSnippet(snippet string) string {
	for line := range strings.SplitSeq(snippet, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, "=") {
			continue
		}
		eq := strings.IndexByte(trimmed, '=')
		if eq <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(trimmed[:eq]))
		value := strings.TrimSpace(trimmed[eq+1:])
		if key == "target_host" || key == "remote" || key == "host" {
			return value
		}
	}
	return ""
}

func communityEndpointKey(item CommunityInterface) string {
	host := strings.ToLower(strings.TrimSpace(item.Host))
	if host == "" {
		host = strings.ToLower(hostFromConfigSnippet(item.Config))
	}
	port := 0
	if item.Port != nil {
		port = *item.Port
	}
	if port == 0 {
		port = portFromConfigSnippet(item.Config)
	}
	if host == "" {
		return "name:" + strings.ToLower(strings.TrimSpace(item.Name))
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func portFromConfigSnippet(snippet string) int {
	for line := range strings.SplitSeq(snippet, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, "=") {
			continue
		}
		eq := strings.IndexByte(trimmed, '=')
		if eq <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(trimmed[:eq]))
		value := strings.TrimSpace(trimmed[eq+1:])
		if key == "target_port" || key == "port" {
			n, err := strconv.Atoi(value)
			if err == nil {
				return n
			}
		}
	}
	return 0
}

func dedupeCommunityInterfaces(items []CommunityInterface) []CommunityInterface {
	seen := make(map[string]bool, len(items))
	out := make([]CommunityInterface, 0, len(items))
	for _, item := range items {
		key := communityEndpointKey(item)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out
}

func ApplyCommunityInterfacesToConfig(cfg *common.ReticulumConfig, items []CommunityInterface) []string {
	if cfg == nil {
		return nil
	}
	added := make([]string, 0, len(items))
	for _, item := range items {
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
			added = append(added, name)
		}
	}
	return added
}

func SeedCommunityInterfaces(cfg *common.ReticulumConfig, items []CommunityInterface, count int) []string {
	picked := PickSeedableCommunityInterfaces(items, count)
	return ApplyCommunityInterfacesToConfig(cfg, picked)
}

func ConfigHasOutboundCommunityInterfaces(cfg *common.ReticulumConfig) bool {
	if cfg == nil {
		return false
	}
	for _, iface := range cfg.Interfaces {
		if iface == nil || !iface.Enabled {
			continue
		}
		t := strings.ToLower(strings.TrimSpace(iface.Type))
		if t == "tcpclientinterface" || strings.Contains(t, "tcpclient") {
			return true
		}
		if strings.Contains(t, "backbone") || strings.Contains(t, "pipe") || strings.Contains(t, "i2p") {
			return true
		}
	}
	return false
}

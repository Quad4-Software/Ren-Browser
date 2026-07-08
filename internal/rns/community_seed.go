// SPDX-License-Identifier: MIT
package rns

import (
	"math/rand"
	"strings"

	"quad4/reticulum-go/pkg/common"
)

const DefaultCommunityInterfaceCount = 4

func PickSeedableCommunityInterfaces(items []CommunityInterface, count int) []CommunityInterface {
	seedable := FilterSeedableInterfaces(items)
	if len(seedable) == 0 {
		return nil
	}
	rand.Shuffle(len(seedable), func(i, j int) { seedable[i], seedable[j] = seedable[j], seedable[i] })
	if count > 0 && len(seedable) > count {
		seedable = seedable[:count]
	}
	return seedable
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
		if usesBackboneTCPFallback() && strings.Contains(t, "backbone") {
			return true
		}
	}
	return false
}

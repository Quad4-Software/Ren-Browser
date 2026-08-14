// SPDX-License-Identifier: MIT
package rns

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed data/community_directory.json
var bundledCommunityDirectory []byte

type CommunityInterface struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	TypeName  string `json:"typeName"`
	Network   string `json:"network"`
	Host      string `json:"host"`
	Port      *int   `json:"port"`
	Status    string `json:"status"`
	Config    string `json:"config"`
	Installed bool   `json:"installed"`
}

type communityDirectoryResponse struct {
	Data []CommunityInterface `json:"data"`
}

func FetchCommunityInterfaces(installed map[string]bool) ([]CommunityInterface, error) {
	return loadBundledCommunityInterfaces(installed)
}

func loadBundledCommunityInterfaces(installed map[string]bool) ([]CommunityInterface, error) {
	var payload communityDirectoryResponse
	if err := json.Unmarshal(bundledCommunityDirectory, &payload); err != nil {
		return nil, fmt.Errorf("decode bundled directory: %w", err)
	}
	return markInstalledCommunityItems(payload.Data, installed), nil
}

func markInstalledCommunityItems(items []CommunityInterface, installed map[string]bool) []CommunityInterface {
	out := make([]CommunityInterface, 0, len(items))
	for _, item := range items {
		if item.Config == "" {
			continue
		}
		item.Installed = installed != nil && installed[item.Name]
		out = append(out, item)
	}
	return out
}

func FilterTCPClientInterfaces(items []CommunityInterface) []CommunityInterface {
	out := make([]CommunityInterface, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Config) == "" {
			continue
		}
		if !IsTCPClientInterface(item) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func IsTCPClientInterface(item CommunityInterface) bool {
	if strings.TrimSpace(item.Config) == "" {
		return false
	}
	t := strings.ToLower(strings.TrimSpace(item.Type))
	tn := strings.ToLower(strings.TrimSpace(item.TypeName))
	if t == "tcpclientinterface" || strings.Contains(t, "tcpclient") {
		return true
	}
	return strings.Contains(tn, "tcp") && strings.Contains(tn, "client")
}

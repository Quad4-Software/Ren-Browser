// SPDX-License-Identifier: MIT
package rns

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"renbrowser/internal/brand"
	"renbrowser/internal/constants"
)

//go:embed data/community_directory.json
var bundledCommunityDirectory []byte

var communityDirectoryURL = constants.CommunityDirectoryURL

func resolveCommunityDirectoryURL() string {
	if v := strings.TrimSpace(os.Getenv(brand.EnvPrefix + "_COMMUNITY_DIRECTORY_URL")); v != "" {
		return v
	}
	return communityDirectoryURL
}

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

type CommunityFetchResult struct {
	Items      []CommunityInterface
	FromBundle bool
}

func FetchCommunityInterfaces(installed map[string]bool) (CommunityFetchResult, error) {
	items, err := fetchLiveCommunityInterfaces(installed)
	if err == nil {
		return CommunityFetchResult{Items: items, FromBundle: false}, nil
	}
	bundled, berr := loadBundledCommunityInterfaces(installed)
	if berr != nil || len(bundled) == 0 {
		return CommunityFetchResult{}, err
	}
	return CommunityFetchResult{Items: bundled, FromBundle: true}, nil
}

func fetchLiveCommunityInterfaces(installed map[string]bool) ([]CommunityInterface, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(resolveCommunityDirectoryURL())
	if err != nil {
		return nil, fmt.Errorf("fetch directory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("directory HTTP %d: %s", resp.StatusCode, string(body))
	}

	var payload communityDirectoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode directory: %w", err)
	}
	return markInstalledCommunityItems(payload.Data, installed), nil
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

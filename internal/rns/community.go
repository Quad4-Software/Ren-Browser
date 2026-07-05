package rns

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const communityDirectoryURL = "https://directory.rns.recipes/api/directory/submitted?search=&type=&status=online"

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
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(communityDirectoryURL)
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

	out := make([]CommunityInterface, 0, len(payload.Data))
	for _, item := range payload.Data {
		if item.Config == "" {
			continue
		}
		item.Installed = installed != nil && installed[item.Name]
		out = append(out, item)
	}
	return out, nil
}

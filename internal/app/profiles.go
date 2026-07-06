// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"renbrowser/internal/nomadnet"
	"renbrowser/internal/store"
)

type ProfileData struct {
	Version        int             `json:"version"`
	Profile        string          `json:"profile"`
	Tabs           []TabSnapshot   `json:"tabs"`
	Favorites      []string        `json:"favorites"`
	History        []HistoryEntry  `json:"history"`
	Keybinds       KeybindSettings `json:"keybinds"`
	BrowserPrefs   BrowserPrefs    `json:"browserPrefs"`
	Theme          ThemeSettings   `json:"theme"`
	DownloadDir    string          `json:"downloadDir,omitempty"`
	Nodes          []nomadnet.Node `json:"nodes,omitempty"`
	WindowState    WindowState     `json:"windowState"`
	EnabledPlugins []string        `json:"enabledPlugins,omitempty"`
}

const profileDataVersion = 1

func (s *BrowserService) ProfileName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profileName
}

func (s *BrowserService) ProfilePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.storePath
}

func (s *BrowserService) ListProfiles() []string {
	names, err := store.ListProfileNames()
	if err != nil {
		return []string{"default"}
	}
	return names
}

func (s *BrowserService) collectProfileData() (ProfileData, error) {
	history, err := s.store.BrowsingHistory(500)
	if err != nil {
		history = []HistoryEntry{}
	}
	nodes, err := s.store.ListNodes()
	if err != nil {
		nodes = []nomadnet.Node{}
	}
	downloadDir := s.GetDownloadDir()
	windowState, _ := s.loadWindowState()
	enabledPlugins := []string{}
	if mgr := s.PluginManager(); mgr != nil {
		enabledPlugins = mgr.EnabledIDs()
	}
	return ProfileData{
		Version:        profileDataVersion,
		Profile:        s.ProfileName(),
		Tabs:           s.store.Tabs(),
		Favorites:      s.store.Favorites(),
		History:        history,
		Keybinds:       s.GetKeybinds(),
		BrowserPrefs:   s.GetBrowserPrefs(),
		Theme:          s.GetTheme(),
		DownloadDir:    downloadDir,
		Nodes:          nodes,
		WindowState:    windowState,
		EnabledPlugins: enabledPlugins,
	}, nil
}

func (s *BrowserService) ExportProfile() (string, error) {
	data, err := s.collectProfileData()
	if err != nil {
		return "", err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (s *BrowserService) ImportProfile(jsonData string) error {
	var data ProfileData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return err
	}
	return s.applyProfileData(data)
}

func (s *BrowserService) applyProfileData(data ProfileData) error {
	if len(data.Tabs) > 0 {
		_ = s.store.SaveTabs(data.Tabs)
	}
	for _, fav := range data.Favorites {
		_ = s.store.AddFavorite(fav)
	}
	for _, entry := range data.History {
		_ = s.store.AddHistory(entry.URL, entry.Title, entry.NodeHash)
	}
	if data.Keybinds.Bindings != nil {
		_ = s.store.SetSetting(keybindsSettingKey, mustEncodeKeybinds(data.Keybinds))
	}
	encoded, err := encodeBrowserPrefs(mergeBrowserPrefs(data.BrowserPrefs))
	if err == nil {
		_ = s.store.SetSetting(browserPrefsKey, encoded)
	}
	if data.DownloadDir != "" {
		_ = s.SetDownloadDir(data.DownloadDir)
	}
	for _, node := range data.Nodes {
		_ = s.store.UpsertNode(node)
	}
	if data.Theme.Mode != "" {
		s.SetTheme(data.Theme)
	}
	if data.WindowState.Width > 0 && data.WindowState.Height > 0 {
		_ = s.saveWindowState(data.WindowState)
	}
	return nil
}

func mustEncodeKeybinds(settings KeybindSettings) string {
	encoded, err := encodeKeybinds(mergeKeybinds(settings))
	if err != nil {
		return ""
	}
	return encoded
}

func (s *BrowserService) exportProfileToFile(path string) error {
	jsonData, err := s.ExportProfile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(jsonData), 0o600)
}

func (s *BrowserService) importProfileFromFile(path string) error {
	raw, err := os.ReadFile(path) // #nosec G304 -- user-supplied profile import path
	if err != nil {
		return err
	}
	return s.ImportProfile(string(raw))
}

func (s *BrowserService) runStartupProfileIO(exportPath, importPath string) error {
	if importPath != "" {
		if err := s.importProfileFromFile(importPath); err != nil {
			return err
		}
	}
	if exportPath != "" {
		if err := s.exportProfileToFile(exportPath); err != nil {
			return err
		}
	}
	return nil
}

func validateProfileImport(jsonData string) error {
	var data ProfileData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return err
	}
	if data.Version > profileDataVersion {
		return errors.New("unsupported profile version")
	}
	return nil
}

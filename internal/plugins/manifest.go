// SPDX-License-Identifier: MIT
package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"renbrowser/internal/buildinfo"
)

const (
	ManifestFileName   = "renbrowser.plugin.json"
	CurrentManifestVer = 1
	MaxManifestBytes   = 256 * 1024
)

type Manifest struct {
	ManifestVersion int               `json:"manifestVersion"`
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description,omitempty"`
	Author          string            `json:"author,omitempty"`
	License         string            `json:"license,omitempty"`
	Engines         map[string]string `json:"engines,omitempty"`
	Main            string            `json:"main,omitempty"`
	Backend         string            `json:"backend,omitempty"`
	Permissions     []string          `json:"permissions,omitempty"`
	Network         *PluginNetwork    `json:"network,omitempty"`
	Contributes     Contributions     `json:"contributes"`
}

type PluginNetwork struct {
	Endpoints []string `json:"endpoints,omitempty"`
}

type Contributions struct {
	Renderers  []RendererContrib  `json:"renderers,omitempty"`
	URLSchemes []URLSchemeContrib `json:"urlSchemes,omitempty"`
	Panels     []PanelContrib     `json:"panels,omitempty"`
	Commands   []CommandContrib   `json:"commands,omitempty"`
	Themes     []ThemeContrib     `json:"themes,omitempty"`
	Settings   []SettingsContrib  `json:"settings,omitempty"`
	DevTools   []DevToolsContrib  `json:"devtools,omitempty"`
}

type RendererContrib struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Extensions []string `json:"extensions,omitempty"`
	MIME       []string `json:"mime,omitempty"`
	Priority   int      `json:"priority,omitempty"`
}

type URLSchemeContrib struct {
	Scheme  string `json:"scheme"`
	Handler string `json:"handler,omitempty"`
}

type PanelContrib struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Icon     string `json:"icon,omitempty"`
	Entry    string `json:"entry"`
	Location string `json:"location,omitempty"`
}

type CommandContrib struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Keybind string `json:"keybind,omitempty"`
}

type ThemeContrib struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

type SettingsContrib struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Entry string `json:"entry"`
}

type DevToolsContrib struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Entry string `json:"entry"`
}

func LoadManifest(dir string) (Manifest, error) {
	path := filepath.Join(dir, ManifestFileName)
	raw, err := os.ReadFile(path) // #nosec G304 -- plugin dir from user data
	if err != nil {
		return Manifest{}, err
	}
	if len(raw) > MaxManifestBytes {
		return Manifest{}, fmt.Errorf("manifest too large")
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	if err := ValidateManifest(m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

func ValidateManifest(m Manifest) error {
	if m.ManifestVersion == 0 {
		m.ManifestVersion = 1
	}
	if m.ManifestVersion > CurrentManifestVer {
		return fmt.Errorf("unsupported manifestVersion %d", m.ManifestVersion)
	}
	if strings.TrimSpace(m.ID) == "" {
		return fmt.Errorf("manifest id is required")
	}
	if !validPluginID(m.ID) {
		return fmt.Errorf("invalid plugin id %q", m.ID)
	}
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("manifest name is required")
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("manifest version is required")
	}
	if err := checkEngine(m.Engines); err != nil {
		return err
	}
	for _, p := range m.Contributes.Panels {
		if p.ID == "" || p.Title == "" || p.Entry == "" {
			return fmt.Errorf("panel contributions require id, title, and entry")
		}
	}
	for _, s := range m.Contributes.URLSchemes {
		if strings.TrimSpace(s.Scheme) == "" {
			return fmt.Errorf("url scheme contribution requires scheme")
		}
	}
	return nil
}

func validPluginID(id string) bool {
	if len(id) < 3 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '.' || r == '-':
		default:
			return false
		}
	}
	return true
}

func checkEngine(engines map[string]string) error {
	constraint, ok := engines["renbrowser"]
	if !ok || strings.TrimSpace(constraint) == "" {
		return nil
	}
	constraint = strings.TrimSpace(constraint)
	if !strings.HasPrefix(constraint, ">=") {
		return fmt.Errorf("unsupported engine constraint %q", constraint)
	}
	want := strings.TrimPrefix(constraint, ">=")
	if !semverGTE(buildinfo.Version, want) {
		return fmt.Errorf("requires renbrowser %s (running %s)", constraint, buildinfo.Version)
	}
	return nil
}

func semverGTE(current, minimum string) bool {
	if current == "dev" || minimum == "dev" {
		return true
	}
	return normalizeSemver(current) >= normalizeSemver(minimum)
}

func normalizeSemver(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	for len(parts) < 3 {
		parts = append(parts, "0")
	}
	return strings.Join(parts[:3], ".")
}

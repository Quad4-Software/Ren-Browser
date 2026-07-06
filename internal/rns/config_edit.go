// SPDX-License-Identifier: MIT
package rns

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/reticulumconfig"

	"renbrowser/internal/brand"
	"renbrowser/internal/paths"
)

func ReadConfigText(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("config path not set")
	}
	data, err := os.ReadFile(path) // #nosec G304 -- reticulum config path from app settings
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func SaveConfigText(path, text string) (*common.ReticulumConfig, error) {
	if path == "" {
		return nil, fmt.Errorf("config path not set")
	}
	cfg, err := loadConfigFromText(text)
	if err != nil {
		return nil, err
	}
	cfg.ConfigPath = filepath.Clean(path)
	if err := reticulumconfig.SaveConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Stack) ReloadConfigFile() error {
	path := s.ConfigPath()
	if path == "" {
		return fmt.Errorf("config path not set")
	}
	cfg, err := loadConfig(path)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
	if !s.started {
		return nil
	}
	return s.ReloadInterfaces(cfg)
}

func (s *Stack) ImportInterfaceConfigs(snippets []string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cfg == nil {
		return nil, errConfigNotLoaded
	}

	added := make([]string, 0)
	for _, snippet := range snippets {
		if strings.TrimSpace(snippet) == "" {
			continue
		}
		ifaces, err := parseInterfaceFragment(snippet)
		if err != nil {
			return added, err
		}
		for name, iface := range ifaces {
			if iface == nil {
				continue
			}
			if _, exists := s.cfg.Interfaces[name]; exists {
				continue
			}
			iface.Name = name
			s.cfg.Interfaces[name] = iface
			added = append(added, name)
		}
	}

	if len(added) == 0 {
		return added, nil
	}
	if err := reticulumconfig.SaveConfig(s.cfg); err != nil {
		return nil, err
	}
	if !s.started {
		return added, nil
	}
	if err := s.ReloadInterfaces(s.cfg); err != nil {
		return added, err
	}
	return added, nil
}

func (s *Stack) InstalledInterfaceNames() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.installedInterfaceNames()
}

func (s *Stack) ReplaceConfig(cfg *common.ReticulumConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
}

func (s *Stack) ApplyConfig(cfg *common.ReticulumConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
	if !s.started {
		return nil
	}
	return s.ReloadInterfaces(cfg)
}

func (s *Stack) installedInterfaceNames() map[string]bool {
	names := make(map[string]bool)
	if s.cfg == nil {
		return names
	}
	for name := range s.cfg.Interfaces {
		names[name] = true
	}
	return names
}

// configTempDir returns a writable scratch directory for the temporary files
// loadConfigFromText needs. The OS default temp dir (os.CreateTemp("", ...))
// is not writable inside the Android app sandbox, so use a directory under
// the app's own data root instead.
func configTempDir() (string, error) {
	dir := filepath.Join(paths.DataRoot(), ".reticulum-go", "tmp")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

func loadConfigFromText(text string) (*common.ReticulumConfig, error) {
	dir, err := configTempDir()
	if err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(dir, brand.TempFilePrefix+"*.conf")
	if err != nil {
		return nil, err
	}
	path := tmp.Name()
	defer os.Remove(path)

	if _, err := tmp.WriteString(text); err != nil {
		_ = tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	return reticulumconfig.LoadConfig(path)
}

func parseInterfaceFragment(snippet string) (map[string]*common.InterfaceConfig, error) {
	normalized := normalizeConfigSnippet(snippet)
	text := "[interfaces]\n\n" + normalized + "\n"
	cfg, err := loadConfigFromText(text)
	if err != nil {
		return nil, err
	}
	if len(cfg.Interfaces) == 0 {
		return nil, fmt.Errorf("no interface found in snippet")
	}
	return cfg.Interfaces, nil
}

func normalizeConfigSnippet(snippet string) string {
	lines := strings.Split(strings.ReplaceAll(snippet, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "remote") {
			if eq := strings.IndexByte(trimmed, '='); eq > 0 {
				value := strings.TrimSpace(trimmed[eq+1:])
				out = append(out, fmt.Sprintf("    target_host = %s", value))
				continue
			}
		}
		if strings.HasPrefix(trimmed, "[[") && strings.HasSuffix(trimmed, "]]") {
			out = append(out, trimmed)
			continue
		}
		if strings.Contains(trimmed, "=") {
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				out = append(out, "    "+trimmed)
				continue
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

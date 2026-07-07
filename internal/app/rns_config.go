// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"os"
	"path/filepath"

	"renbrowser/internal/rns"
)

func (s *BrowserService) ShowConfigDir() error {
	path := s.ConfigPath()
	if path == "" {
		return fmt.Errorf("config path not set")
	}
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("config directory not found")
	}
	return platformOpenPath(dir)
}

func (s *BrowserService) GetReticulumConfigText() (string, error) {
	path := s.ConfigPath()
	if path == "" {
		return "", fmt.Errorf("reticulum not initialized")
	}
	return rns.ReadConfigText(path)
}

func (s *BrowserService) SaveReticulumConfigText(text string) error {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return fmt.Errorf("reticulum not initialized")
	}
	path := stack.ConfigPath()
	cfg, err := rns.SaveConfigText(path, text)
	if err != nil {
		return err
	}
	if err := stack.ApplyConfig(cfg); err != nil {
		return err
	}
	s.log("info", "reticulum config saved", path)
	if s.app != nil {
		s.app.Event.Emit("rns:status", "reload")
	}
	return nil
}

func (s *BrowserService) ReloadReticulumConfig() (string, error) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return "", fmt.Errorf("reticulum not initialized")
	}
	if err := stack.ReloadConfigFile(); err != nil {
		return "", err
	}
	path := stack.ConfigPath()
	text, err := rns.ReadConfigText(path)
	if err != nil {
		return "", err
	}
	s.log("info", "reticulum config reloaded", path)
	if s.app != nil {
		s.app.Event.Emit("rns:status", "reload")
	}
	return text, nil
}

func (s *BrowserService) ImportCommunityInterfaces(configs []string) ([]string, error) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return nil, fmt.Errorf("reticulum not initialized")
	}
	added, err := stack.ImportInterfaceConfigs(configs)
	if err != nil {
		return added, err
	}
	if len(added) > 0 {
		s.log("info", "community interfaces added", fmt.Sprintf("%v", added))
		if s.app != nil {
			s.app.Event.Emit("rns:status", "reload")
		}
	}
	return added, nil
}

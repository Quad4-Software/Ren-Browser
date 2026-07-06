// SPDX-License-Identifier: MIT
package app

import "github.com/wailsapp/wails/v3/pkg/application"

type RuntimeConfig struct {
	PublicMode  bool   `json:"publicMode"`
	ServerMode  bool   `json:"serverMode"`
	Profile     string `json:"profile"`
	ProfilePath string `json:"profilePath"`
}

func (s *BrowserService) GetRuntimeConfig() RuntimeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return RuntimeConfig{
		PublicMode:  s.publicMode,
		ServerMode:  application.System.IsServer(),
		Profile:     s.profileName,
		ProfilePath: s.storePath,
	}
}

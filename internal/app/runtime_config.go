// SPDX-License-Identifier: MIT
package app

type RuntimeConfig struct {
	PublicMode  bool   `json:"publicMode"`
	Profile     string `json:"profile"`
	ProfilePath string `json:"profilePath"`
}

func (s *BrowserService) GetRuntimeConfig() RuntimeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return RuntimeConfig{
		PublicMode:  s.publicMode,
		Profile:     s.profileName,
		ProfilePath: s.storePath,
	}
}

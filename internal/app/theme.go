// SPDX-License-Identifier: MIT
package app

import "maps"

import "encoding/json"

const themeSettingKey = "theme"

func DefaultThemeSettings() ThemeSettings {
	return ThemeSettings{
		Mode:            "dark",
		Accent:          "#60a5fa",
		FontFamily:      "system-ui, -apple-system, Segoe UI, sans-serif",
		FontSize:        14,
		CustomTokens:    map[string]string{},
		CompactToolbar:  false,
		OverlaySidebars: false,
	}
}

func mergeThemeSettings(saved ThemeSettings) ThemeSettings {
	defaults := DefaultThemeSettings()
	if saved.Mode != "" {
		defaults.Mode = saved.Mode
	}
	if saved.Accent != "" {
		defaults.Accent = saved.Accent
	}
	if saved.FontFamily != "" {
		defaults.FontFamily = saved.FontFamily
	}
	if saved.FontSize > 0 {
		defaults.FontSize = saved.FontSize
	}
	if saved.CustomTokens != nil {
		tokens := make(map[string]string, len(saved.CustomTokens))
		maps.Copy(tokens, saved.CustomTokens)
		defaults.CustomTokens = tokens
	}
	defaults.CompactToolbar = saved.CompactToolbar
	defaults.OverlaySidebars = saved.OverlaySidebars
	return defaults
}

func encodeThemeSettings(theme ThemeSettings) (string, error) {
	raw, err := json.Marshal(theme)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func decodeThemeSettings(raw string) (ThemeSettings, error) {
	if raw == "" {
		return DefaultThemeSettings(), nil
	}
	var theme ThemeSettings
	if err := json.Unmarshal([]byte(raw), &theme); err != nil {
		return DefaultThemeSettings(), err
	}
	return mergeThemeSettings(theme), nil
}

func (s *BrowserService) loadThemeSettings() ThemeSettings {
	raw, err := s.store.GetSetting(themeSettingKey)
	if err != nil {
		return DefaultThemeSettings()
	}
	theme, err := decodeThemeSettings(raw)
	if err != nil {
		return DefaultThemeSettings()
	}
	return theme
}

func (s *BrowserService) GetTheme() ThemeSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.theme
}

func (s *BrowserService) SetTheme(theme ThemeSettings) ThemeSettings {
	theme = mergeThemeSettings(theme)
	s.mu.Lock()
	s.theme = theme
	s.mu.Unlock()
	if encoded, err := encodeThemeSettings(theme); err == nil {
		_ = s.store.SetSetting(themeSettingKey, encoded)
	}
	if s.app != nil {
		s.app.Event.Emit("theme:changed", theme)
	}
	return theme
}

// SPDX-License-Identifier: MIT
package app

import "encoding/json"

const browserPrefsKey = "browserPrefs"

type BrowserPrefs struct {
	OpenLinksInNewTab         bool            `json:"openLinksInNewTab"`
	OpenLinksInNewWindow      bool            `json:"openLinksInNewWindow"`
	NativeTitlebar            bool            `json:"nativeTitlebar"`
	MicronRenderer            string          `json:"micronRenderer"`
	MicronWasmEnabled         bool            `json:"micronWasmEnabled"`
	MicronWasmParserID        string          `json:"micronWasmParserId"`
	DocsLanguage              string          `json:"docsLanguage"`
	UILanguage                string          `json:"uiLanguage"`
	DiscoverySlowMode         bool            `json:"discoverySlowMode"`
	MobileDevTools            bool            `json:"mobileDevTools"`
	PageCacheEnabled          bool            `json:"pageCacheEnabled"`
	TabHoverPreviews          bool            `json:"tabHoverPreviews"`
	SettingsSectionsCollapsed map[string]bool `json:"settingsSectionsCollapsed"`
}

func DefaultBrowserPrefs() BrowserPrefs {
	return BrowserPrefs{
		OpenLinksInNewTab:  true,
		MicronRenderer:     "auto",
		MicronWasmEnabled:  true,
		MicronWasmParserID: "bundled",
		PageCacheEnabled:   true,
		TabHoverPreviews:   true,
	}
}

func mergeBrowserPrefs(saved BrowserPrefs) BrowserPrefs {
	defaults := DefaultBrowserPrefs()
	defaults.OpenLinksInNewTab = saved.OpenLinksInNewTab
	defaults.OpenLinksInNewWindow = saved.OpenLinksInNewWindow
	defaults.NativeTitlebar = saved.NativeTitlebar
	if saved.MicronRenderer != "" {
		defaults.MicronRenderer = saved.MicronRenderer
	}
	defaults.MicronWasmEnabled = saved.MicronWasmEnabled
	if saved.MicronWasmParserID != "" {
		defaults.MicronWasmParserID = saved.MicronWasmParserID
	}
	if saved.DocsLanguage != "" {
		defaults.DocsLanguage = saved.DocsLanguage
	}
	if saved.UILanguage != "" {
		defaults.UILanguage = saved.UILanguage
	}
	defaults.DiscoverySlowMode = saved.DiscoverySlowMode
	defaults.MobileDevTools = saved.MobileDevTools
	defaults.PageCacheEnabled = saved.PageCacheEnabled
	defaults.TabHoverPreviews = saved.TabHoverPreviews
	if len(saved.SettingsSectionsCollapsed) > 0 {
		defaults.SettingsSectionsCollapsed = saved.SettingsSectionsCollapsed
	}
	return defaults
}

func encodeBrowserPrefs(prefs BrowserPrefs) (string, error) {
	raw, err := json.Marshal(prefs)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func decodeBrowserPrefs(raw string) (BrowserPrefs, error) {
	if raw == "" {
		return DefaultBrowserPrefs(), nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return DefaultBrowserPrefs(), err
	}
	var prefs BrowserPrefs
	if err := json.Unmarshal([]byte(raw), &prefs); err != nil {
		return DefaultBrowserPrefs(), err
	}
	merged := mergeBrowserPrefs(prefs)
	if _, ok := fields["micronWasmEnabled"]; !ok {
		merged.MicronWasmEnabled = DefaultBrowserPrefs().MicronWasmEnabled
	}
	if _, ok := fields["pageCacheEnabled"]; !ok {
		merged.PageCacheEnabled = DefaultBrowserPrefs().PageCacheEnabled
	}
	if _, ok := fields["tabHoverPreviews"]; !ok {
		merged.TabHoverPreviews = DefaultBrowserPrefs().TabHoverPreviews
	}
	return merged, nil
}

func (s *BrowserService) GetBrowserPrefs() BrowserPrefs {
	raw, err := s.store.GetSetting(browserPrefsKey)
	if err != nil {
		return DefaultBrowserPrefs()
	}
	prefs, err := decodeBrowserPrefs(raw)
	if err != nil {
		return DefaultBrowserPrefs()
	}
	return prefs
}

func (s *BrowserService) SetBrowserPrefs(prefs BrowserPrefs) BrowserPrefs {
	merged := mergeBrowserPrefs(prefs)
	encoded, err := encodeBrowserPrefs(merged)
	if err != nil {
		return DefaultBrowserPrefs()
	}
	_ = s.store.SetSetting(browserPrefsKey, encoded)
	return merged
}

// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"strings"
)

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
	MicronPreserveLayout      bool            `json:"micronPreserveLayout"`
	InitialSetupComplete      bool            `json:"initialSetupComplete"`
	SettingsSectionsCollapsed map[string]bool `json:"settingsSectionsCollapsed"`
	// MicronImagesMode controls inline image loading on micron pages:
	// "off" hides the load control, "ask" shows a per-image opt-in button,
	// "always" auto-loads. Nodes listed in MicronImageNodes override the
	// global mode with "always" or "never".
	MicronImagesMode string            `json:"micronImagesMode"`
	MicronImageNodes map[string]string `json:"micronImageNodes"`
}

func DefaultBrowserPrefs() BrowserPrefs {
	return BrowserPrefs{
		OpenLinksInNewTab:  true,
		NativeTitlebar:     platformDefaultNativeTitlebar(),
		MicronRenderer:     "auto",
		MicronWasmEnabled:  true,
		MicronWasmParserID: "bundled",
		PageCacheEnabled:   true,
		TabHoverPreviews:   true,
		MicronImagesMode:   "ask",
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
	defaults.MicronPreserveLayout = saved.MicronPreserveLayout
	defaults.InitialSetupComplete = saved.InitialSetupComplete
	if len(saved.SettingsSectionsCollapsed) > 0 {
		defaults.SettingsSectionsCollapsed = saved.SettingsSectionsCollapsed
	}
	switch saved.MicronImagesMode {
	case "off", "ask", "always":
		defaults.MicronImagesMode = saved.MicronImagesMode
	}
	if len(saved.MicronImageNodes) > 0 {
		merged := make(map[string]string, len(saved.MicronImageNodes))
		for hash, policy := range saved.MicronImageNodes {
			key := strings.ToLower(strings.TrimSpace(hash))
			if !nodeImageHashRe.MatchString(key) {
				continue
			}
			if policy != "always" && policy != "never" {
				continue
			}
			merged[key] = policy
		}
		if len(merged) > 0 {
			defaults.MicronImageNodes = merged
		}
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
	if _, ok := fields["nativeTitlebar"]; !ok {
		merged.NativeTitlebar = DefaultBrowserPrefs().NativeTitlebar
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

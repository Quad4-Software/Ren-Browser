// SPDX-License-Identifier: MIT
package app

type SettingsReset struct {
	Theme        ThemeSettings   `json:"theme"`
	Keybinds     KeybindSettings `json:"keybinds"`
	BrowserPrefs BrowserPrefs    `json:"browserPrefs"`
}

func (s *BrowserService) ResetSettings() SettingsReset {
	keybinds := DefaultKeybinds()
	_ = s.store.SetSetting(keybindsSettingKey, "")
	_ = s.store.SetSetting(themeSettingKey, "")
	prefs := DefaultBrowserPrefs()
	encoded, err := encodeBrowserPrefs(prefs)
	if err == nil {
		_ = s.store.SetSetting(browserPrefsKey, encoded)
	}
	theme := s.SetTheme(DefaultThemeSettings())
	if s.app != nil {
		window := s.app.Window.Current()
		if window != nil {
			window.SetFrameless(true)
		}
		s.app.Event.Emit("window:chrome", WindowChrome{NativeTitlebar: false})
	}
	return SettingsReset{
		Theme:        theme,
		Keybinds:     keybinds,
		BrowserPrefs: prefs,
	}
}

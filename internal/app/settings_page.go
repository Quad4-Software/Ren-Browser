// SPDX-License-Identifier: MIT
package app

import "strings"

func isSettingsURL(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "settings", "settings:":
		return true
	default:
		return false
	}
}

func (s *BrowserService) settingsPage(pushHistory bool) PageResponse {
	resp := PageResponse{
		URL:         "settings:",
		Path:        "/settings",
		ContentType: "settings",
	}
	if pushHistory {
		s.pushHistory("settings:")
		_ = s.store.AddHistory("settings:", "Settings", "")
	}
	s.setLastPage(resp)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
	}
	return resp
}

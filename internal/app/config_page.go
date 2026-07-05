// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"strings"

	"renbrowser/internal/rns"
)

func isConfigURL(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "config", "config:":
		return true
	default:
		return false
	}
}

func (s *BrowserService) configPage(pushHistory bool) PageResponse {
	raw := ""
	path := s.ConfigPath()
	if path != "" {
		text, err := rns.ReadConfigText(path)
		if err != nil {
			raw = fmt.Sprintf("# failed to read config: %v", err)
		} else {
			raw = text
		}
	}
	resp := PageResponse{
		URL:         "config:",
		Path:        "/config",
		ContentType: "config",
		Raw:         raw,
	}
	if pushHistory {
		s.pushHistory("config:")
		_ = s.store.AddHistory("config:", "Reticulum Config", "")
	}
	s.setLastPage(resp)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
	}
	return resp
}

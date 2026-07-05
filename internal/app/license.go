// SPDX-License-Identifier: MIT
package app

import (
	"strings"

	"renbrowser/internal/content"
)

func isLicenseURL(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "license", "license:":
		return true
	default:
		return false
	}
}

func (s *BrowserService) licensePage(pushHistory bool) PageResponse {
	html := content.RenderLicense()
	resp := PageResponse{
		URL:         "license:",
		Path:        "/license",
		ContentType: "license",
		HTML:        html,
		Raw:         html,
	}
	if pushHistory {
		s.pushHistory("license:")
		_ = s.store.AddHistory("license:", "License", "")
	}
	s.setLastPage(resp)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
	}
	return resp
}

// SPDX-License-Identifier: MIT
package builtin

import (
	"strings"

	"renbrowser/internal/plugins"
)

type SchemeDeps struct {
	AboutHTML    func() string
	EditorRaw    func() string
	ConfigRaw    func() string
	LicenseHTML  func() string
	DocsPage     func(rawURL string) (plugins.SchemeResult, bool)
	SettingsPage func() (plugins.SchemeResult, bool)
}

func RegisterSchemes(reg *plugins.Registry, deps SchemeDeps) {
	reg.RegisterScheme("", plugins.URLSchemeContrib{Scheme: "about"}, func(rawURL string) (plugins.SchemeResult, bool) {
		if !matchAbout(rawURL) {
			return plugins.SchemeResult{}, false
		}
		html := ""
		if deps.AboutHTML != nil {
			html = deps.AboutHTML()
		}
		return plugins.SchemeResult{
			URL:          "about:",
			Path:         "/about",
			ContentType:  "about",
			HTML:         html,
			Raw:          html,
			HistoryTitle: "About",
		}, true
	})
	reg.RegisterScheme("", plugins.URLSchemeContrib{Scheme: "editor"}, func(rawURL string) (plugins.SchemeResult, bool) {
		if !matchEditor(rawURL) {
			return plugins.SchemeResult{}, false
		}
		raw := ""
		if deps.EditorRaw != nil {
			raw = deps.EditorRaw()
		}
		return plugins.SchemeResult{
			URL:          "editor:",
			Path:         "/page/editor.mu",
			ContentType:  "editor",
			Raw:          raw,
			HistoryTitle: "Micron Editor",
		}, true
	})
	reg.RegisterScheme("", plugins.URLSchemeContrib{Scheme: "config"}, func(rawURL string) (plugins.SchemeResult, bool) {
		if !matchConfig(rawURL) {
			return plugins.SchemeResult{}, false
		}
		raw := ""
		if deps.ConfigRaw != nil {
			raw = deps.ConfigRaw()
		}
		return plugins.SchemeResult{
			URL:          "config:",
			Path:         "/config",
			ContentType:  "config",
			Raw:          raw,
			HistoryTitle: "Reticulum Config",
		}, true
	})
	reg.RegisterScheme("", plugins.URLSchemeContrib{Scheme: "license"}, func(rawURL string) (plugins.SchemeResult, bool) {
		if !matchLicense(rawURL) {
			return plugins.SchemeResult{}, false
		}
		html := ""
		if deps.LicenseHTML != nil {
			html = deps.LicenseHTML()
		}
		return plugins.SchemeResult{
			URL:          "license:",
			Path:         "/license",
			ContentType:  "license",
			HTML:         html,
			Raw:          html,
			HistoryTitle: "License",
		}, true
	})
	reg.RegisterScheme("", plugins.URLSchemeContrib{Scheme: "docs"}, func(rawURL string) (plugins.SchemeResult, bool) {
		if !matchDocs(rawURL) {
			return plugins.SchemeResult{}, false
		}
		if deps.DocsPage == nil {
			return plugins.SchemeResult{}, false
		}
		return deps.DocsPage(rawURL)
	})
	reg.RegisterScheme("", plugins.URLSchemeContrib{Scheme: "settings"}, func(rawURL string) (plugins.SchemeResult, bool) {
		if !matchSettings(rawURL) {
			return plugins.SchemeResult{}, false
		}
		if deps.SettingsPage == nil {
			return plugins.SchemeResult{}, false
		}
		return deps.SettingsPage()
	})
}

func matchAbout(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "about", "about:":
		return true
	default:
		return false
	}
}

func matchEditor(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "editor", "editor:":
		return true
	default:
		return false
	}
}

func matchConfig(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "config", "config:":
		return true
	default:
		return false
	}
}

func matchLicense(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "license", "license:":
		return true
	default:
		return false
	}
}

func matchDocs(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case lower == "docs", lower == "docs:":
		return true
	case strings.HasPrefix(lower, "docs?"):
		return true
	case strings.HasPrefix(lower, "docs:?"):
		return true
	default:
		return false
	}
}

func matchSettings(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "settings", "settings:":
		return true
	default:
		return false
	}
}

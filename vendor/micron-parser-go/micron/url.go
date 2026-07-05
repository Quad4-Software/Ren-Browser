// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"regexp"
	"strings"
)

var schemePrefix = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`)

// FormatNomadnetworkURL ensures URLs use a scheme Micron and NomadNet tooling expect.
// If url already starts with a letter scheme and "://", it is usually returned as-is.
// Otherwise "nomadnetwork://" is prepended.
// javascript:, vbscript:, and file: (including javascript://) always get that prefix
// so generic HTML embeds never put a raw executable or file URL in href.
func FormatNomadnetworkURL(raw string) string {
	url := strings.TrimSpace(raw)
	if url == "" {
		return ""
	}
	if schemePrefix.MatchString(url) {
		if dangerousNavScheme(url) {
			return "nomadnetwork://" + url
		}
		return url
	}
	return "nomadnetwork://" + url
}

func dangerousNavScheme(url string) bool {
	colon := strings.IndexByte(url, ':')
	if colon <= 0 {
		return false
	}
	scheme := strings.ToLower(url[:colon])
	switch scheme {
	case "javascript", "vbscript", "file":
		return true
	default:
		return false
	}
}

func linkDirectURL(raw string) string {
	return strings.ReplaceAll(strings.ReplaceAll(raw, "nomadnetwork://", ""), "lxmf://", "")
}

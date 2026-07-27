// SPDX-License-Identifier: MIT
package content

import (
	"regexp"
)

var (
	scriptBlockRe = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	scriptOpenRe  = regexp.MustCompile(`(?is)<script[^>]*>?`)
	styleBlockRe  = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	onAttrRe      = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	iframeRe      = regexp.MustCompile(`(?is)<iframe\b[^>]*>.*?</iframe>`)
	objectRe      = regexp.MustCompile(`(?is)<object\b[^>]*>.*?</object>`)
	embedRe       = regexp.MustCompile(`(?is)<embed\b[^>]*/?>`)
	jsSchemeRe    = regexp.MustCompile(`(?i)javascript:`)
	metaRefreshRe = regexp.MustCompile(`(?is)<meta\b[^>]*http-equiv\s*=\s*["']?refresh["']?[^>]*>`)
	baseHrefRe    = regexp.MustCompile(`(?is)<base\b[^>]*>`)
)

func SanitizeHTML(input string) string {
	if input == "" || !needsHTMLSanitize(input) {
		return input
	}
	out := input
	for range 3 {
		if !mayContainScript(out) {
			break
		}
		prev := out
		out = scriptBlockRe.ReplaceAllString(out, "")
		out = scriptOpenRe.ReplaceAllString(out, "")
		if out == prev {
			break
		}
	}
	out = styleBlockRe.ReplaceAllString(out, "")
	out = iframeRe.ReplaceAllString(out, "")
	out = objectRe.ReplaceAllString(out, "")
	out = embedRe.ReplaceAllString(out, "")
	out = metaRefreshRe.ReplaceAllString(out, "")
	out = baseHrefRe.ReplaceAllString(out, "")
	out = onAttrRe.ReplaceAllString(out, "")
	out = jsSchemeRe.ReplaceAllString(out, "")
	return out
}

func needsHTMLSanitize(s string) bool {
	n := len(s)
	for i := range n {
		switch s[i] {
		case 'j', 'J':
			if i+10 < n && asciiEqualFold(s[i:i+11], "javascript:") {
				return true
			}
		case '<':
			if dangerousHTMLTagAt(s, i) {
				return true
			}
		case ' ', '\t', '\n', '\r':
			if i+3 < n && asciiEqualFold(s[i+1:i+3], "on") {
				c := s[i+3]
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
					return true
				}
			}
		}
	}
	return false
}

func dangerousHTMLTagAt(s string, i int) bool {
	if s[i] != '<' {
		return false
	}
	j := i + 1
	if j < len(s) && s[j] == '/' {
		j++
	}
	if j >= len(s) {
		return false
	}
	switch asciiLower(s[j]) {
	case 'b':
		return htmlTagNameAt(s, j, "base")
	case 'e':
		return htmlTagNameAt(s, j, "embed")
	case 'i':
		return htmlTagNameAt(s, j, "iframe")
	case 'm':
		return htmlTagNameAt(s, j, "meta")
	case 'o':
		return htmlTagNameAt(s, j, "object")
	case 's':
		if htmlTagNameAt(s, j, "script") || htmlTagNameAt(s, j, "style") {
			return true
		}
		return looksLikePartialTag(s, j, "script")
	default:
		return false
	}
}

func mayContainScript(s string) bool {
	n := len(s)
	for i := 0; i < n; i++ {
		if s[i] != '<' {
			continue
		}
		j := i + 1
		if j < n && s[j] == '/' {
			j++
		}
		if j+1 < n && asciiEqualFoldByte(s[j], 's') && asciiEqualFoldByte(s[j+1], 'c') {
			return true
		}
	}
	return false
}

func looksLikePartialTag(s string, j int, tag string) bool {
	for k := 0; k < len(tag); k++ {
		if j+k >= len(s) {
			return k > 0
		}
		if !asciiEqualFoldByte(s[j+k], tag[k]) {
			return false
		}
	}
	return true
}

func htmlTagNameAt(s string, j int, tag string) bool {
	if j+len(tag) > len(s) {
		return false
	}
	for k := 0; k < len(tag); k++ {
		if !asciiEqualFoldByte(s[j+k], tag[k]) {
			return false
		}
	}
	if j+len(tag) == len(s) {
		return true
	}
	switch s[j+len(tag)] {
	case '>', ' ', '\t', '\n', '\r', '/':
		return true
	default:
		return false
	}
}

func asciiLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

func asciiEqualFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !asciiEqualFoldByte(a[i], b[i]) {
			return false
		}
	}
	return true
}

func asciiEqualFoldByte(a, b byte) bool {
	if a >= 'A' && a <= 'Z' {
		a += 'a' - 'A'
	}
	if b >= 'A' && b <= 'Z' {
		b += 'a' - 'A'
	}
	return a == b
}

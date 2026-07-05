// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"html"
	"strings"
)

// stripASCIIControls removes ASCII control characters (U+0000–U+001F).
// html.EscapeString does not escape NUL or line breaks. Dropping C0 controls
// keeps visible text and attribute values predictable.
func stripASCIIControls(s string) string {
	if strings.IndexFunc(s, func(r rune) bool { return r < 0x20 }) < 0 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 0x20 {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func htmlText(s string) string {
	return html.EscapeString(stripASCIIControls(s))
}

func appendHTMLText(b *strings.Builder, s string) {
	s = stripASCIIControls(s)
	start := 0
	for i := 0; i < len(s); i++ {
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			esc = "&#34;"
		case '\'':
			esc = "&#39;"
		default:
			continue
		}
		if start < i {
			b.WriteString(s[start:i])
		}
		b.WriteString(esc)
		start = i + 1
	}
	if start < len(s) {
		b.WriteString(s[start:])
	}
}

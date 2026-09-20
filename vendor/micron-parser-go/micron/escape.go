// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"unicode/utf8"
)

// stripASCIIControls removes ASCII control characters (U+0000-U+001F).
// html.EscapeString does not escape NUL or line breaks. Dropping C0 controls
// keeps visible text and attribute values predictable.
func stripASCIIControls(s string) string {
	for i := range len(s) {
		if s[i] >= 0x20 {
			continue
		}
		var b strings.Builder
		b.Grow(len(s))
		b.WriteString(s[:i])
		for j := i; j < len(s); j++ {
			if s[j] >= 0x20 {
				b.WriteByte(s[j])
			}
		}
		return b.String()
	}
	return s
}

// isNerdIconRune returns true for runes in the Private Use Areas that
// Nerd Font / patched icon fonts occupy. These glyphs need a font that
// includes them, otherwise they render as replacement boxes.
func isNerdIconRune(r rune) bool {
	return (r >= 0xE000 && r <= 0xF8FF) ||
		(r >= 0xF0000 && r <= 0xFFFFD) ||
		(r >= 0x100000 && r <= 0x10FFFD)
}

func appendHTMLText(b *strings.Builder, s string) {
	s = stripASCIIControls(s)
	var icon bool
	start := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		nextIcon := isNerdIconRune(r)
		if nextIcon != icon {
			if start < i {
				b.WriteString(s[start:i])
			}
			if nextIcon {
				b.WriteString(`<span class="nf" style="font-family:'Roboto Mono Nerd Font',monospace">`)
			} else {
				b.WriteString(`</span>`)
			}
			icon = nextIcon
			start = i
		}
		if !nextIcon && size == 1 {
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
			}
			if esc != "" {
				if start < i {
					b.WriteString(s[start:i])
				}
				b.WriteString(esc)
				start = i + 1
			}
		}
		i += size
	}
	if start < len(s) {
		b.WriteString(s[start:])
	}
	if icon {
		b.WriteString(`</span>`)
	}
}

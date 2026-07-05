// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"math"
	"strconv"
	"strings"
)

const (
	defaultBG      = "default"
	defaultFGDark  = "ddd"
	defaultFGLight = "222"
)

// ColorToCSS maps Micron color tokens (hex, grayscale gNN, defaults) to CSS color strings.
// It returns "" when c is empty, "default", or not recognized.
func ColorToCSS(c string) string {
	if !micronColorToken(c) {
		return ""
	}
	var b strings.Builder
	b.Grow(8)
	writeMicronColorHex(&b, c)
	return b.String()
}

func micronColorToken(c string) bool {
	if c == "" || c == defaultBG {
		return false
	}
	if len(c) == 3 && isHex3(c) {
		return true
	}
	if len(c) == 6 && isHex6(c) {
		return true
	}
	return len(c) == 3 && c[0] == 'g'
}

func writeMicronColorHex(b *strings.Builder, c string) {
	switch {
	case len(c) == 3 && isHex3(c):
		b.WriteByte('#')
		b.WriteString(c)
	case len(c) == 6 && isHex6(c):
		b.WriteByte('#')
		b.WriteString(c)
	default:
		v, err := strconv.Atoi(c[1:])
		if err != nil || v < 0 {
			v = 50
		}
		if v > 99 {
			v = 99
		}
		h := byte(math.Floor(float64(v) * 2.55))
		const hx = "0123456789abcdef"
		b.WriteByte('#')
		b.WriteByte(hx[h>>4])
		b.WriteByte(hx[h&0xf])
		b.WriteByte(hx[h>>4])
		b.WriteByte(hx[h&0xf])
		b.WriteByte(hx[h>>4])
		b.WriteByte(hx[h&0xf])
	}
}

func tryAppendColorProperty(b *strings.Builder, prop string, c string) bool {
	if !micronColorToken(c) {
		return false
	}
	b.WriteString(prop)
	writeMicronColorHex(b, c)
	return true
}

func isHex3(s string) bool {
	for i := range 3 {
		if !isHexByte(s[i]) {
			return false
		}
	}
	return true
}

func isHex6(s string) bool {
	for i := range 6 {
		if !isHexByte(s[i]) {
			return false
		}
	}
	return true
}

func isHexByte(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func appendStyleAttr(b *strings.Builder, st Style, defaultBG string) {
	if tryAppendColorProperty(b, "color:", st.FG) {
		b.WriteByte(';')
	}
	if st.BG != defaultBG && st.BG != "default" && tryAppendColorProperty(b, "background-color:", st.BG) {
		b.WriteString(";display:inline-block;")
	}
	if st.Bold {
		b.WriteString("font-weight:bold;")
	}
	if st.Underline {
		b.WriteString("text-decoration:underline;")
	}
	if st.Italic {
		b.WriteString("font-style:italic;")
	}
}

func styleAttr(st Style, defaultBG string) string {
	var b strings.Builder
	b.Grow(96)
	appendStyleAttr(&b, st, defaultBG)
	return b.String()
}

func headingStyle(p *Parser, level int) Style {
	if p.DarkTheme {
		switch level {
		case 1:
			return Style{FG: "222", BG: "bbb", Bold: false, Underline: false, Italic: false}
		case 2:
			return Style{FG: "111", BG: "999", Bold: false, Underline: false, Italic: false}
		case 3:
			return Style{FG: "000", BG: "777", Bold: false, Underline: false, Italic: false}
		}
		return plainStyle(p)
	}
	switch level {
	case 1:
		return Style{FG: "000", BG: "777", Bold: false, Underline: false, Italic: false}
	case 2:
		return Style{FG: "111", BG: "aaa", Bold: false, Underline: false, Italic: false}
	case 3:
		return Style{FG: "222", BG: "ccc", Bold: false, Underline: false, Italic: false}
	}
	return plainStyle(p)
}

func plainStyle(p *Parser) Style {
	fg := defaultFGLight
	if p.DarkTheme {
		fg = defaultFGDark
	}
	return Style{FG: fg, BG: defaultBG, Bold: false, Underline: false, Italic: false}
}

func (p *Parser) stateToStyle(s *State) Style {
	return Style{
		FG:        s.FGColor,
		BG:        s.BGColor,
		Bold:      s.Formatting.Bold,
		Underline: s.Formatting.Underline,
		Italic:    s.Formatting.Italic,
	}
}

func (p *Parser) styleToState(st Style, s *State) {
	s.FGColor = st.FG
	s.BGColor = st.BG
	s.Formatting.Bold = st.Bold
	s.Formatting.Underline = st.Underline
	s.Formatting.Italic = st.Italic
}

func stylesEqual(a, b *Style) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.FG == b.FG && a.BG == b.BG && a.Bold == b.Bold &&
		a.Underline == b.Underline && a.Italic == b.Italic
}

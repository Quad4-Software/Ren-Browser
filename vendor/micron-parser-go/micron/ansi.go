// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strconv"
	"strings"
)

// RenderANSI renders a Document to a simple ANSI / plain terminal string.
// It is intended for TUI hosts (NomadNet-class) rather than HTML.
func (p *Parser) RenderANSI(doc *Document) string {
	if doc == nil {
		return ""
	}
	var b strings.Builder
	b.Grow(len(doc.Source) + 64)
	for i := range doc.Blocks {
		bl := &doc.Blocks[i]
		switch bl.Kind {
		case BlockBlank:
			b.WriteByte('\n')
		case BlockHR:
			b.WriteString(strings.Repeat("-", 40))
			b.WriteByte('\n')
		case BlockDivider:
			r := bl.DividerRune
			if r == 0 {
				r = '*'
			}
			b.WriteString(strings.Repeat(string(r), 40))
			b.WriteByte('\n')
		case BlockHeading:
			indent := strings.Repeat("  ", max(bl.Depth-1, 0))
			b.WriteString(indent)
			b.WriteString(ansiStyleOpen(bl.HeadingStyle))
			writeInlinesPlain(&b, bl.Inlines)
			b.WriteString(ansiReset())
			b.WriteByte('\n')
		case BlockParagraph:
			indent := strings.Repeat("  ", max(bl.Depth-1, 0))
			b.WriteString(indent)
			writeInlinesPlain(&b, bl.Inlines)
			b.WriteByte('\n')
		case BlockPartial:
			if bl.Partial != nil {
				b.WriteString("[partial:")
				b.WriteString(bl.Partial.Destination)
				b.WriteString("]\n")
			}
		}
	}
	return b.String()
}

func writeInlinesPlain(b *strings.Builder, inlines []Inline) {
	for i := range inlines {
		in := &inlines[i]
		if in.Field != nil {
			b.WriteString("[")
			b.WriteString(in.Field.Name)
			b.WriteByte('=')
			if in.Field.Kind == FieldCheckbox || in.Field.Kind == FieldRadio {
				b.WriteString(in.Field.Label)
			} else {
				b.WriteString(in.Field.Value)
			}
			b.WriteByte(']')
			continue
		}
		if in.Link != nil {
			b.WriteString(stripHTMLRough(in.Link.Label))
			b.WriteString(" <")
			b.WriteString(in.Link.URL)
			b.WriteByte('>')
			continue
		}
		if in.Partial != nil {
			b.WriteString("[partial:")
			b.WriteString(in.Partial.Destination)
			b.WriteByte(']')
			continue
		}
		b.WriteString(ansiStyleOpen(in.Style))
		b.WriteString(in.Text)
		if hasANSIStyle(in.Style) {
			b.WriteString(ansiReset())
		}
	}
}

func hasANSIStyle(st Style) bool {
	return st.Bold || st.Italic || st.Underline || micronColorToken(st.FG) || micronColorToken(st.BG)
}

func ansiStyleOpen(st Style) string {
	if !hasANSIStyle(st) {
		return ""
	}
	var b strings.Builder
	b.WriteString("\x1b[")
	first := true
	writeCode := func(code string) {
		if !first {
			b.WriteByte(';')
		}
		first = false
		b.WriteString(code)
	}
	if st.Bold {
		writeCode("1")
	}
	if st.Italic {
		writeCode("3")
	}
	if st.Underline {
		writeCode("4")
	}
	if css := ColorToCSS(st.FG); len(css) == 7 && css[0] == '#' {
		writeCode("38;2;" + rgbCSV(css))
	}
	if css := ColorToCSS(st.BG); len(css) == 7 && css[0] == '#' {
		writeCode("48;2;" + rgbCSV(css))
	}
	if first {
		return ""
	}
	b.WriteByte('m')
	return b.String()
}

func ansiReset() string { return "\x1b[0m" }

func rgbCSV(cssHex string) string {
	// #rgb or #rrggbb already expanded by ColorToCSS to #rrggbb or #rgb
	h := cssHex[1:]
	if len(h) == 3 {
		r := hexNibble(h[0]) * 17
		g := hexNibble(h[1]) * 17
		b := hexNibble(h[2]) * 17
		return strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b)
	}
	if len(h) != 6 {
		return "255;255;255"
	}
	r := hexNibble(h[0])*16 + hexNibble(h[1])
	g := hexNibble(h[2])*16 + hexNibble(h[3])
	b := hexNibble(h[4])*16 + hexNibble(h[5])
	return strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b)
}

func hexNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return 0
	}
}

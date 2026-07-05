// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

// ConvertMicronToHTML renders Micron markup to a self-contained HTML fragment.
// Text is escaped and ASCII control characters (U+0000–U+001F) are stripped from
// emitted text and attributes. Only parser-emitted tags and attributes appear in the output.
// The caller supplies the full document. Optional leading #!fg= / #!bg= lines set default colors.
// Treat the result as safe HTML only together with a sensible CSP and link handling policy on the host.
func (p *Parser) ConvertMicronToHTML(markup string) string {
	pc := ParseHeaderTags(markup)
	plain := plainStyle(p)
	defaultFG := pc.FG
	if defaultFG == "" {
		defaultFG = plain.FG
	}
	defaultBGVal := plain.BG
	if pc.BG != "" {
		defaultBGVal = pc.BG
	}
	s := State{
		Literal:      false,
		Depth:        0,
		FGColor:      defaultFG,
		BGColor:      defaultBGVal,
		DefaultAlign: "left",
		Align:        "left",
		DefaultFG:    defaultFG,
		DefaultBG:    defaultBGVal,
	}
	var b strings.Builder
	if len(markup) > 0 {
		// Output is often larger than input. Pre-grow reduces reallocations for typical docs.
		b.Grow(4 * len(markup))
	}
	for start := 0; start <= len(markup); {
		nextRel := strings.IndexByte(markup[start:], '\n')
		line := ""
		if nextRel < 0 {
			line = markup[start:]
			start = len(markup) + 1
		} else {
			next := start + nextRel
			line = markup[start:next]
			start = next + 1
		}
		k := p.parseLineInto(&b, line, &s)
		switch k {
		case lineOmit:
			continue
		}
	}
	var out strings.Builder
	out.Grow(b.Len() + 128)
	out.WriteString(`<div style="line-height:1.5;`)
	if defaultFG != "" && defaultFG != "default" && tryAppendColorProperty(&out, "color:", defaultFG) {
		out.WriteByte(';')
	}
	if defaultBGVal != "" && defaultBGVal != "default" && tryAppendColorProperty(&out, "background-color:", defaultBGVal) {
		out.WriteByte(';')
	}
	out.WriteString(`">`)
	out.WriteString(b.String())
	out.WriteString(`</div>`)
	return out.String()
}

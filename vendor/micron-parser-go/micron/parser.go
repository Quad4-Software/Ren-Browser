// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

// ConvertMicronToHTML renders Micron markup to a self-contained HTML fragment.
// Text is escaped and ASCII control characters (U+0000-U+001F) are stripped from
// emitted text and attributes. Only parser-emitted tags and attributes appear in the output.
// The caller supplies the full document. Optional leading #!fg= / #!bg= lines set default colors.
// Treat the result as safe HTML only together with a sensible CSP and link handling policy on the host.
//
// Convert streams HTML in one pass without building a Document IR. Use Parse and
// RenderHTML when you need the AST, spans, or diagnostics.
func (p *Parser) ConvertMicronToHTML(markup string) string {
	pc := ParseHeaderTags(markup)
	return p.streamHTML(markup, pc)
}

// ConvertMicronToHTMLWithColors renders markup and returns page colors from leading directives.
func (p *Parser) ConvertMicronToHTMLWithColors(markup string) (html, fg, bg string) {
	pc := ParseHeaderTags(markup)
	html = p.streamHTML(markup, pc)
	return html, pc.FG, pc.BG
}

func (p *Parser) streamHTML(markup string, pc PageColors) string {
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
		DefaultAlign: "start",
		Align:        "start",
		DefaultFG:    defaultFG,
		DefaultBG:    defaultBGVal,
		FoldOpen:     pc.FoldOpen,
		FoldClosed:   pc.FoldClosed,
	}
	if s.FoldOpen == "" {
		s.FoldOpen = "▾"
	}
	if s.FoldClosed == "" {
		s.FoldClosed = "▸"
	}
	var b strings.Builder
	n := len(markup)
	if n > 0 {
		if p.ForceMonospace {
			b.Grow(8*n + 160)
		} else {
			b.Grow(4*n + 160)
		}
	} else {
		b.Grow(160)
	}
	writeRootOpen(&b, defaultFG, defaultBGVal)
	lineNum := 1
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
		p.parseLineInto(&b, line, &s, lineNum)
		lineNum++
	}
	p.closeFolds(&b, &s, 0)
	b.WriteString(`</div>`)
	return b.String()
}

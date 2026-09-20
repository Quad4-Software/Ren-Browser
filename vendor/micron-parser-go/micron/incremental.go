// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

// ParseIncremental reparses markup after an edit described by oldLen/newSegment.
// When the change is confined to a suffix of lines, unchanged leading blocks are
// reused from prev. If prev is nil or reuse is unsafe, it falls back to Parse.
func (p *Parser) ParseIncremental(prev *Document, markup string) *Document {
	if prev == nil || prev.Source == "" {
		return p.Parse(markup)
	}
	old := prev.Source
	// Find longest common line-aligned prefix.
	oi, ni := 0, 0
	lastNL := 0
	for oi < len(old) && ni < len(markup) && old[oi] == markup[ni] {
		if old[oi] == '\n' {
			lastNL = oi + 1
		}
		oi++
		ni++
	}
	// If divergence is not at a line boundary, reparse from previous newline.
	prefixEnd := lastNL
	if prefixEnd == 0 && oi > 0 && oi == ni && oi == len(old) && ni == len(markup) {
		return prev
	}
	// Count complete blocks fully within the common prefix byte range.
	reuse := 0
	for reuse < len(prev.Blocks) {
		bl := prev.Blocks[reuse]
		// Require a strict end-before-prefix so a trailing blank at the
		// common boundary is not kept when the suffix continues.
		if bl.Span.End >= prefixEnd {
			break
		}
		reuse++
	}
	if reuse == 0 || prefixEnd == 0 {
		return p.Parse(markup)
	}
	suffix := markup[prefixEnd:]
	suffixDoc := p.Parse(suffix)
	lineBase := 1
	for i := range prefixEnd {
		if markup[i] == '\n' {
			lineBase++
		}
	}
	out := &Document{
		Source:    markup,
		Colors:    prev.Colors,
		DefaultFG: prev.DefaultFG,
		DefaultBG: prev.DefaultBG,
		Blocks:    make([]Block, 0, reuse+len(suffixDoc.Blocks)),
	}
	// Prefer header colors from a full reparse of headers if prefix includes them.
	if strings.HasPrefix(markup, "#!") {
		full := p.Parse(markup)
		out.Colors = full.Colors
		out.DefaultFG = full.DefaultFG
		out.DefaultBG = full.DefaultBG
	}
	out.Blocks = append(out.Blocks, prev.Blocks[:reuse]...)
	for i := range suffixDoc.Blocks {
		bl := suffixDoc.Blocks[i]
		bl.Span.Start += prefixEnd
		bl.Span.End += prefixEnd
		if bl.SourceLine > 0 {
			bl.SourceLine += lineBase - 1
		}
		for j := range bl.Inlines {
			bl.Inlines[j].Span.Start += prefixEnd
			bl.Inlines[j].Span.End += prefixEnd
		}
		out.Blocks = append(out.Blocks, bl)
	}
	return out
}

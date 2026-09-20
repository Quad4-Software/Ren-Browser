// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// RenderHTML renders a Document to a self-contained HTML fragment.
// Block wrappers include data-mu-line with the 1-based source line for editor correlation.
func (p *Parser) RenderHTML(doc *Document) string {
	if doc == nil {
		plain := plainStyle(p)
		return p.wrapHTMLFragment("", plain.FG, plain.BG)
	}
	s := State{
		Literal:      false,
		Depth:        0,
		FGColor:      doc.DefaultFG,
		BGColor:      doc.DefaultBG,
		DefaultAlign: "start",
		Align:        "start",
		DefaultFG:    doc.DefaultFG,
		DefaultBG:    doc.DefaultBG,
	}
	var b strings.Builder
	if doc.Source != "" {
		b.Grow(4*len(doc.Source) + 160)
	} else {
		b.Grow(160)
	}
	writeRootOpen(&b, doc.DefaultFG, doc.DefaultBG)
	for i := range doc.Blocks {
		p.renderBlock(&b, &doc.Blocks[i], &s)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func (p *Parser) wrapHTMLFragment(body, defaultFG, defaultBGVal string) string {
	var out strings.Builder
	out.Grow(len(body) + 160)
	writeRootOpen(&out, defaultFG, defaultBGVal)
	out.WriteString(body)
	out.WriteString(`</div>`)
	return out.String()
}

func writeRootOpen(b *strings.Builder, defaultFG, defaultBGVal string) {
	b.WriteString(`<div class="Mu-root" dir="auto" style="line-height:1.5;min-height:100%;width:100%;box-sizing:border-box;`)
	if defaultFG != "" && defaultFG != "default" && tryAppendColorProperty(b, "color:", defaultFG) {
		b.WriteByte(';')
	}
	if defaultBGVal != "" && defaultBGVal != "default" && tryAppendColorProperty(b, "background-color:", defaultBGVal) {
		b.WriteByte(';')
	}
	b.WriteString(`">`)
}

func writeDataMuLine(b *strings.Builder, line int) {
	if line <= 0 {
		return
	}
	b.WriteString(` data-mu-line="`)
	b.WriteString(strconv.Itoa(line))
	b.WriteByte('"')
}

func (p *Parser) renderBlock(out *strings.Builder, bl *Block, s *State) {
	s.Depth = bl.Depth
	s.Align = bl.Align
	if s.Align == "" {
		s.Align = "start"
	}
	line := bl.SourceLine
	switch bl.Kind {
	case BlockBlank:
		s.BGColor = bl.LineBG
		if s.BGColor != s.DefaultBG && s.BGColor != "default" && micronColorToken(s.BGColor) {
			out.WriteString(`<div`)
			writeDataMuLine(out, line)
			out.WriteString(` style="background-color:`)
			writeMicronColorHex(out, s.BGColor)
			out.WriteString(`;width:100%;display:block;height:1.2em;"><div style="`)
			appendSectionIndentStyleNoSemi(out, s)
			out.WriteString(`"><br></div></div>`)
			return
		}
		out.WriteString(`<br`)
		writeDataMuLine(out, line)
		out.WriteByte('>')
	case BlockPartial:
		if bl.Partial != nil {
			p.writePartial(out, bl.Partial, s, line)
		}
	case BlockHR:
		s.FGColor = bl.HeadingStyle.FG
		s.BGColor = bl.HeadingStyle.BG
		out.WriteString(`<hr`)
		writeDataMuLine(out, line)
		out.WriteString(` style="all:revert;`)
		if tryAppendColorProperty(out, "border-color:", s.FGColor) {
			out.WriteByte(';')
		}
		out.WriteString(`margin:0.5em 0 0.5em 0;`)
		if s.BGColor != s.DefaultBG && s.BGColor != "default" && micronColorToken(s.BGColor) {
			out.WriteString(`box-shadow:0 0 0 0.5em `)
			writeMicronColorHex(out, s.BGColor)
			out.WriteByte(';')
		}
		appendSectionIndentStyle(out, s)
		out.WriteString(`"/>`)
	case BlockDivider:
		s.FGColor = bl.HeadingStyle.FG
		s.BGColor = bl.HeadingStyle.BG
		out.WriteString(`<div`)
		writeDataMuLine(out, line)
		out.WriteString(` style="white-space:pre;white-space:nowrap;overflow:hidden;width:100%;`)
		if tryAppendColorProperty(out, "color:", s.FGColor) {
			out.WriteByte(';')
		}
		if s.BGColor != s.DefaultBG && s.BGColor != "default" && tryAppendColorProperty(out, "background-color:", s.BGColor) {
			out.WriteByte(';')
		}
		appendSectionIndentStyle(out, s)
		out.WriteString(`">`)
		var tmp [utf8.UTFMax]byte
		n := utf8.EncodeRune(tmp[:], bl.DividerRune)
		rText := string(tmp[:n])
		for range 250 {
			appendHTMLText(out, rText)
		}
		out.WriteString(`</div>`)
	case BlockHeading:
		style := bl.HeadingStyle
		out.WriteString(`<div`)
		writeDataMuLine(out, line)
		out.WriteString(` style="display:block;width:100%;`)
		if tryAppendColorProperty(out, "color:", style.FG) {
			out.WriteByte(';')
		}
		if tryAppendColorProperty(out, "background-color:", style.BG) {
			out.WriteByte(';')
		}
		out.WriteString(`"><div style="`)
		appendSectionIndentStyle(out, s)
		out.WriteString(`">`)
		s.Literal = false
		p.appendOutput(out, inlinesToParts(bl.Inlines), s)
		out.WriteString(`</div></div><br>`)
	case BlockParagraph:
		s.BGColor = bl.LineBG
		s.Literal = bl.Literal
		if !p.ForceMonospace && bl.Literal && len(bl.Inlines) == 1 && bl.Inlines[0].Field == nil && bl.Inlines[0].Link == nil && bl.Inlines[0].Partial == nil {
			p.appendWrappedAlignedFastPlain(out, bl.Inlines[0].Text, s, line)
			return
		}
		p.appendWrappedAlignedParts(out, inlinesToParts(bl.Inlines), s, line)
	}
}

func inlinesToParts(inlines []Inline) []linePart {
	if len(inlines) == 0 {
		return nil
	}
	parts := make([]linePart, len(inlines))
	for i := range inlines {
		in := &inlines[i]
		parts[i] = linePart{
			style:   in.Style,
			text:    in.Text,
			field:   in.Field,
			link:    in.Link,
			partial: in.Partial,
		}
	}
	return parts
}

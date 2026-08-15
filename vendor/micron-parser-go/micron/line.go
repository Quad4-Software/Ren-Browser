// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	lineOmit = iota
	lineHTML
)

// isLiteralToggleLine reports whether the line is exactly a "`=" toggle
// surrounded only by ASCII whitespace, without allocating. Matches
// micron-parser-js / NomadNet "line.trim() === '`='" without a TrimSpace
// substring scan past the toggle.
func isLiteralToggleLine(line string) bool {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t' || line[i] == '\r') {
		i++
	}
	if i+2 > len(line) || line[i] != '`' || line[i+1] != '=' {
		return false
	}
	j := i + 2
	for j < len(line) {
		c := line[j]
		if c != ' ' && c != '\t' && c != '\r' {
			return false
		}
		j++
	}
	return true
}

// trimASCIISpaces trims ASCII spaces, tabs, and \r from both ends of s. It is
// a fast-path for the common case of micron lines that only carry ASCII
// whitespace; sufficient because input is already split on \n.
func trimASCIISpaces(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r') {
		i++
	}
	j := len(s)
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}

func (p *Parser) parseLineInto(out *strings.Builder, line string, s *State) int {
	if len(line) > 0 {
		if isLiteralToggleLine(line) {
			s.Literal = !s.Literal
			return lineOmit
		}
		preEscape := false
		if !s.Literal {
			if line[0] == '>' && strings.Contains(line, "`<") {
				k := 0
				for k < len(line) && line[k] == '>' {
					k++
				}
				line = line[k:]
				if len(line) == 0 {
					return p.parseLineInto(out, "", s)
				}
			}
			if line[0] == '\\' {
				line = line[1:]
				preEscape = true
			} else if line[0] == '#' {
				return lineOmit
			} else if len(line) >= 2 && line[0] == '`' && line[1] == 't' {
				return p.consumeTableFence(out, line, s)
			} else if s.TableMode {
				s.TableLines = append(s.TableLines, line)
				return lineOmit
			} else if len(line) >= 2 && line[0] == '`' && line[1] == '{' {
				pt := p.parsePartialFromInner(line[2:], s)
				if pt == nil {
					return lineOmit
				}
				p.writePartial(out, pt, s)
				return lineHTML
			} else if line[0] == '<' {
				s.Depth = 0
				if len(line) == 1 {
					return lineOmit
				}
				return p.parseLineInto(out, line[1:], s)
			} else if line[0] == '>' {
				i := 0
				for i < len(line) && line[i] == '>' {
					i++
				}
				s.Depth = i
				headingLine := trimASCIISpaces(line[i:])
				if headingLine == "" {
					return lineOmit
				}
				style := headingStyle(p, i)
				latched := p.stateToStyle(s)
				p.styleToState(style, s)
				parts := p.makeOutput(s, headingLine, false)
				p.styleToState(latched, s)
				if !partsHaveContent(parts) {
					p.styleToState(latched, s)
					return lineOmit
				}
				out.WriteString(`<div style="display:inline-block;width:100%;`)
				if tryAppendColorProperty(out, "color:", style.FG) {
					out.WriteByte(';')
				}
				if tryAppendColorProperty(out, "background-color:", style.BG) {
					out.WriteByte(';')
				}
				out.WriteString(`"><div style="`)
				appendSectionIndentStyle(out, s)
				out.WriteString(`">`)
				p.appendOutput(out, parts, s)
				out.WriteString(`</div></div><br>`)
				return lineHTML
			} else if line[0] == '-' {
				if len(line) == 1 {
					out.WriteString(`<hr style="all:revert;`)
					if tryAppendColorProperty(out, "border-color:", s.FGColor) {
						out.WriteByte(';')
					}
					out.WriteString(`margin:0.5em 0 0.5em 0;`)
					if micronColorToken(s.BGColor) {
						out.WriteString(`box-shadow:0 0 0 0.5em `)
						writeMicronColorHex(out, s.BGColor)
						out.WriteByte(';')
					}
					appendSectionIndentStyle(out, s)
					out.WriteString(`"/>`)
					return lineHTML
				}
				_, firstSize := utf8.DecodeRuneInString(line)
				r, _ := utf8.DecodeRuneInString(line[firstSize:])
				out.WriteString(`<div style="white-space:pre;white-space:nowrap;overflow:hidden;width:100%;`)
				if tryAppendColorProperty(out, "color:", s.FGColor) {
					out.WriteByte(';')
				}
				if s.BGColor != s.DefaultBG && s.BGColor != "default" && tryAppendColorProperty(out, "background-color:", s.BGColor) {
					out.WriteByte(';')
				}
				appendSectionIndentStyle(out, s)
				out.WriteString(`">`)
				var tmp [utf8.UTFMax]byte
				n := utf8.EncodeRune(tmp[:], r)
				rText := string(tmp[:n])
				for range 250 {
					appendHTMLText(out, rText)
				}
				out.WriteString(`</div>`)
				return lineHTML
			}
		}
		if !s.Literal && strings.IndexByte(line, '`') < 0 && !preEscape {
			parts := p.makeOutput(s, line, false)
			p.appendWrappedAlignedParts(out, parts, s)
			return lineHTML
		}
		if !p.ForceMonospace && s.Literal {
			text := line
			if line == "\\`=" {
				text = "`="
			}
			p.appendWrappedAlignedFastPlain(out, text, s)
			return lineHTML
		}
		parts := p.makeOutput(s, line, preEscape)
		p.appendWrappedAlignedParts(out, parts, s)
		return lineHTML
	}
	if s.BGColor != s.DefaultBG && s.BGColor != "default" && micronColorToken(s.BGColor) {
		out.WriteString(`<div style="background-color:`)
		writeMicronColorHex(out, s.BGColor)
		out.WriteString(`;width:100%;display:block;height:1.2em;"><div style="`)
		appendSectionIndentStyleNoSemi(out, s)
		out.WriteString(`"><br></div></div>`)
		return lineHTML
	}
	out.WriteString(`<br>`)
	return lineHTML
}

// partsHaveContent reports whether any element in parts will produce visible
// HTML output. Used to skip empty wrapper divs without round-tripping through
// an intermediate strings.Builder.
func partsHaveContent(parts []linePart) bool {
	for i := range parts {
		pr := &parts[i]
		if pr.field != nil || pr.link != nil || pr.partial != nil {
			return true
		}
		if pr.html != "" || pr.text != "" {
			return true
		}
	}
	return false
}

func (p *Parser) appendWrappedAlignedParts(out *strings.Builder, parts []linePart, s *State) {
	if !partsHaveContent(parts) {
		return
	}
	bg := s.BGColor != s.DefaultBG && s.BGColor != "default" && micronColorToken(s.BGColor)
	if bg {
		out.WriteString(`<div style="background-color:`)
		writeMicronColorHex(out, s.BGColor)
		out.WriteString(`;width:100%;display:block;">`)
	}
	out.WriteString(`<div style="text-align:`)
	out.WriteString(s.Align)
	out.WriteString(`;`)
	appendSectionIndentStyle(out, s)
	out.WriteString(`">`)
	p.appendOutput(out, parts, s)
	out.WriteString(`</div>`)
	if bg {
		out.WriteString(`</div>`)
	}
}

func (p *Parser) appendWrappedAlignedFastPlain(out *strings.Builder, line string, s *State) {
	bg := s.BGColor != s.DefaultBG && s.BGColor != "default" && micronColorToken(s.BGColor)
	if bg {
		out.WriteString(`<div style="background-color:`)
		writeMicronColorHex(out, s.BGColor)
		out.WriteString(`;width:100%;display:block;">`)
	}
	out.WriteString(`<div style="text-align:`)
	out.WriteString(s.Align)
	out.WriteString(`;`)
	appendSectionIndentStyle(out, s)
	out.WriteString(`">`)
	sa := cachedStateStyleAttr(s)
	if sa != "" {
		out.WriteString(`<span style="`)
		out.WriteString(sa)
		out.WriteString(`">`)
	}
	if p.ForceMonospace {
		p.appendSplitAtSpaces(out, line)
	} else {
		appendHTMLText(out, line)
	}
	if sa != "" {
		out.WriteString(`</span>`)
	}
	out.WriteString(`</div>`)
	if bg {
		out.WriteString(`</div>`)
	}
}

func cachedStateStyleAttr(s *State) string {
	key := stateStyleKey{
		FG:        s.FGColor,
		BG:        s.BGColor,
		Bold:      s.Formatting.Bold,
		Underline: s.Formatting.Underline,
		Italic:    s.Formatting.Italic,
	}
	if s.styleAttrMap != nil {
		if v, ok := s.styleAttrMap[key]; ok {
			return v
		}
	} else {
		s.styleAttrMap = make(map[stateStyleKey]string, 8)
	}
	v := styleAttr(Style{
		FG:        key.FG,
		BG:        key.BG,
		Bold:      key.Bold,
		Underline: key.Underline,
		Italic:    key.Italic,
	}, s.DefaultBG)
	s.styleAttrMap[key] = v
	return v
}

func sectionIndentStyleEm(s *State) float64 {
	ind := max((s.Depth-1)*2, 0)
	if ind <= 0 {
		return 0
	}
	return float64(ind) * 0.6
}

func appendSectionIndentStyle(b *strings.Builder, s *State) {
	em := sectionIndentStyleEm(s)
	if em <= 0 {
		return
	}
	b.WriteString("margin-left:")
	b.WriteString(strconv.FormatFloat(em, 'f', 1, 64))
	b.WriteString("em;")
}

func appendSectionIndentStyleNoSemi(b *strings.Builder, s *State) {
	em := sectionIndentStyleEm(s)
	if em <= 0 {
		return
	}
	b.WriteString("margin-left:")
	b.WriteString(strconv.FormatFloat(em, 'f', 1, 64))
	b.WriteString("em")
}

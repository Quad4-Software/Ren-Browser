// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
)

func backslashEscapesOnlyMicronSpecial(line string, pos int) bool {
	if pos >= len(line) {
		return false
	}
	switch line[pos] {
	case '`', '\\', '[':
		return true
	default:
		return false
	}
}

func (p *Parser) makeOutput(s *State, line string, preEscape bool) []linePart {
	if s.Literal {
		if line == "\\`=" {
			line = "`="
		}
		st := p.stateToStyle(s)
		return []linePart{{style: st, text: line}}
	}

	if strings.IndexByte(line, '`') < 0 && !preEscape {
		st := p.stateToStyle(s)
		return []linePart{{style: st, text: line}}
	}

	out := make([]linePart, 0, 8)
	var part strings.Builder
	modeText := true
	escape := preEscape
	skip := 0
	i := 0

	flushPart := func() {
		if part.Len() == 0 {
			return
		}
		partStr := part.String()
		part.Reset()
		st := p.stateToStyle(s)
		out = append(out, linePart{style: st, text: partStr})
	}

	for i < len(line) {
		if skip > 0 {
			skip--
			i++
			continue
		}
		if !modeText {
			c := line[i]
			if c == '\\' {
				if backslashEscapesOnlyMicronSpecial(line, i+1) {
					modeText = true
					escape = true
					i++
					continue
				}
				part.WriteByte('\\')
				modeText = true
				i++
				continue
			}
			switch c {
			case '_':
				s.Formatting.Underline = !s.Formatting.Underline
			case '!':
				s.Formatting.Bold = !s.Formatting.Bold
			case '*':
				s.Formatting.Italic = !s.Formatting.Italic
			case 'F':
				if i+1 < len(line) && line[i+1] == 'T' && len(line) >= i+8 {
					s.FGColor = line[i+2 : i+8]
					skip = 7
				} else if len(line) >= i+9 && line[i+4] == '`' && line[i+5] == 'F' {
					var b [6]byte
					b[0] = line[i+6]
					b[1] = line[i+1]
					b[2] = line[i+7]
					b[3] = line[i+2]
					b[4] = line[i+8]
					b[5] = line[i+3]
					s.FGColor = string(b[:])
					skip = 8
				} else if len(line) >= i+4 {
					s.FGColor = line[i+1 : i+4]
					skip = 3
				}
			case 'f':
				s.FGColor = s.DefaultFG
			case 'B':
				if i+1 < len(line) && line[i+1] == 'T' && len(line) >= i+8 {
					s.BGColor = line[i+2 : i+8]
					skip = 7
					flushPart()
				} else if len(line) >= i+9 && line[i+4] == '`' && line[i+5] == 'B' {
					var b [6]byte
					b[0] = line[i+6]
					b[1] = line[i+1]
					b[2] = line[i+7]
					b[3] = line[i+2]
					b[4] = line[i+8]
					b[5] = line[i+3]
					s.BGColor = string(b[:])
					skip = 8
					flushPart()
				} else if len(line) >= i+4 {
					s.BGColor = line[i+1 : i+4]
					skip = 3
					flushPart()
				}
			case 'b':
				s.BGColor = s.DefaultBG
				flushPart()
			case '`':
				s.Formatting.Bold = false
				s.Formatting.Underline = false
				s.Formatting.Italic = false
				s.FGColor = s.DefaultFG
				s.BGColor = s.DefaultBG
				s.Align = s.DefaultAlign
				modeText = true
			case 'c':
				s.Align = "center"
			case 'l':
				s.Align = "left"
			case 'r':
				s.Align = "right"
			case 'a':
				s.Align = s.DefaultAlign
			case '<':
				flushPart()
				if sk, f := p.parseField(line, i, s); f != nil {
					out = append(out, linePart{field: f})
					i += sk
					modeText = true
					continue
				}
			case '[':
				flushPart()
				if sk, lk := p.parseLink(line, i, s); lk != nil {
					out = append(out, linePart{link: lk})
					i += sk
					modeText = true
					continue
				}
			}
			modeText = true
			i++
			continue
		}

		c := line[i]
		if escape {
			part.WriteByte(c)
			escape = false
			i++
			continue
		}
		if c == '\\' {
			if !backslashEscapesOnlyMicronSpecial(line, i+1) {
				part.WriteByte('\\')
				i++
				continue
			}
			escape = true
			i++
			continue
		}
		if c == '`' {
			if i+1 < len(line) && line[i+1] == '`' {
				flushPart()
				s.Formatting.Bold = false
				s.Formatting.Underline = false
				s.Formatting.Italic = false
				s.FGColor = s.DefaultFG
				s.BGColor = s.DefaultBG
				s.Align = s.DefaultAlign
				i += 2
				continue
			}
			flushPart()
			modeText = false
			i++
			continue
		}
		part.WriteByte(c)
		i++
	}
	flushPart()
	return out
}

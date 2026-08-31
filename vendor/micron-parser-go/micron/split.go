// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func splitAfterSpaceSegments(s string) []string {
	if s == "" {
		return []string{""}
	}
	segments := strings.Count(s, " ") + 1
	out := make([]string, 0, segments)
	start := 0
	for start < len(s) {
		rel := strings.IndexByte(s[start:], ' ')
		if rel < 0 {
			out = append(out, s[start:])
			break
		}
		end := start + rel + 1
		out = append(out, s[start:end])
		start = end
	}
	return out
}

func (p *Parser) splitAtSpaces(line string) string {
	var b strings.Builder
	p.appendSplitAtSpaces(&b, line)
	return b.String()
}

func (p *Parser) forceMonospace(line string) string {
	if !p.ForceMonospace {
		return htmlText(line)
	}
	var b strings.Builder
	p.appendForceMonospace(&b, line)
	return b.String()
}

func (p *Parser) appendSplitAtSpaces(b *strings.Builder, line string) {
	if line == "" {
		b.WriteString(`<span class="Mu-mws"></span>`)
		return
	}
	start := 0
	for start < len(line) {
		rel := strings.IndexByte(line[start:], ' ')
		end := len(line)
		if rel >= 0 {
			end = start + rel + 1
		}
		b.WriteString(`<span class="Mu-mws">`)
		p.appendForceMonospace(b, line[start:end])
		b.WriteString(`</span>`)
		start = end
	}
}

// isComplexScriptBase reports scripts that require contextual shaping or strong
// RTL layout. Wrapping each rune in display:inline-block Mu-mnt spans breaks
// Arabic Persian Hebrew Syriac and related joining and bidirectional order.
func isComplexScriptBase(r rune) bool {
	switch {
	case r >= 0x0590 && r <= 0x05FF: // Hebrew
		return true
	case r >= 0x0600 && r <= 0x06FF: // Arabic
		return true
	case r >= 0x0700 && r <= 0x074F: // Syriac
		return true
	case r >= 0x0750 && r <= 0x077F: // Arabic Supplement
		return true
	case r >= 0x0780 && r <= 0x07BF: // Thaana
		return true
	case r >= 0x07C0 && r <= 0x07FF: // NKo
		return true
	case r >= 0x0800 && r <= 0x083F: // Samaritan
		return true
	case r >= 0x0840 && r <= 0x085F: // Mandaic
		return true
	case r >= 0x0860 && r <= 0x086F: // Syriac Supplement
		return true
	case r >= 0x0870 && r <= 0x089F: // Arabic Extended-B
		return true
	case r >= 0x08A0 && r <= 0x08FF: // Arabic Extended-A
		return true
	case r >= 0xFB1D && r <= 0xFB4F: // Hebrew presentation forms
		return true
	case r >= 0xFB50 && r <= 0xFDFF: // Arabic Presentation Forms-A
		return true
	case r >= 0xFE70 && r <= 0xFEFF: // Arabic Presentation Forms-B
		return true
	case r >= 0x1EE00 && r <= 0x1EEFF: // Arabic Mathematical Alphabetic Symbols
		return true
	default:
		return false
	}
}

func isCombiningMark(r rune) bool {
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Me, r)
}

// isJoinControl is ZWNJ ZWJ used inside Arabic Persian and related words.
func isJoinControl(r rune) bool {
	return r == '\u200C' || r == '\u200D'
}

func (p *Parser) appendForceMonospace(b *strings.Builder, line string) {
	line = stripASCIIControls(line)
	for i := 0; i < len(line); {
		r, sz := utf8.DecodeRuneInString(line[i:])
		if isComplexScriptBase(r) {
			j := i + sz
			for j < len(line) {
				r2, sz2 := utf8.DecodeRuneInString(line[j:])
				if isComplexScriptBase(r2) || isCombiningMark(r2) || isJoinControl(r2) {
					j += sz2
					continue
				}
				break
			}
			// Emit the whole shaping run as one text node so letters can join.
			appendHTMLText(b, line[i:j])
			i = j
			continue
		}
		end := i + sz
		for end < len(line) {
			r2, sz2 := utf8.DecodeRuneInString(line[end:])
			if !isCombiningMark(r2) {
				break
			}
			end += sz2
		}
		b.WriteString(`<span class="Mu-mnt">`)
		if end == i+1 && line[i] < utf8.RuneSelf {
			switch line[i] {
			case '&':
				b.WriteString("&amp;")
			case '<':
				b.WriteString("&lt;")
			case '>':
				b.WriteString("&gt;")
			case '"':
				b.WriteString("&#34;")
			case '\'':
				b.WriteString("&#39;")
			default:
				b.WriteByte(line[i])
			}
		} else {
			appendHTMLText(b, line[i:end])
		}
		b.WriteString(`</span>`)
		i = end
	}
}

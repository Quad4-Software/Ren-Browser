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

// appendSplitAtSpaces mirrors micron-parser-js splitAtSpaces / wrapWord.
// Every space-delimited segment is wrapped in a whitespace-preserving span so
// multiple consecutive spaces are not collapsed. Plain printable ASCII words
// (no & < >) use Mu-mnt-group; words that need escaping or contain non-ASCII
// use Mu-mws with Mu-mnt cells around complex grapheme clusters and HTML-special
// bytes.
func (p *Parser) appendSplitAtSpaces(b *strings.Builder, line string) {
	if line == "" {
		b.WriteString(`<span class="Mu-mnt-group"></span>`)
		return
	}
	start := 0
	for start < len(line) {
		rel := strings.IndexByte(line[start:], ' ')
		end := len(line)
		if rel >= 0 {
			end = start + rel + 1
		}
		word := line[start:end]
		if wordNeedsMonoWrap(word) {
			b.WriteString(`<span class="Mu-mws">`)
			p.appendForceMonospace(b, word)
			b.WriteString(`</span>`)
		} else {
			b.WriteString(`<span class="Mu-mnt-group">`)
			b.WriteString(word)
			b.WriteString(`</span>`)
		}
		start = end
	}
}

func wordNeedsMonoWrap(word string) bool {
	for i := range len(word) {
		c := word[i]
		if c < 0x20 || c >= 0x7F || c == '&' || c == '<' || c == '>' {
			return true
		}
	}
	return false
}

// isComplexScriptBase reports scripts that must stay as continuous text runs
// under ForceMonospace. Wrapping each rune in display:inline-block Mu-mnt spans
// breaks Arabic/Persian joining, Hebrew/RTL order, CJK spacing, and Indic/Thai
// cluster shaping.
func isComplexScriptBase(r rune) bool {
	switch {
	case r >= 0x0400 && r <= 0x04FF: // Cyrillic
		return true
	case r >= 0x0500 && r <= 0x052F: // Cyrillic Supplement
		return true
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
	case r >= 0x0900 && r <= 0x097F: // Devanagari
		return true
	case r >= 0x0E00 && r <= 0x0E7F: // Thai
		return true
	case r >= 0x1100 && r <= 0x11FF: // Hangul Jamo
		return true
	case r >= 0x3040 && r <= 0x309F: // Hiragana
		return true
	case r >= 0x30A0 && r <= 0x30FF: // Katakana
		return true
	case r >= 0x3130 && r <= 0x318F: // Hangul Compatibility Jamo
		return true
	case r >= 0x3400 && r <= 0x4DBF: // CJK Unified Ideographs Extension A
		return true
	case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified Ideographs
		return true
	case r >= 0xAC00 && r <= 0xD7AF: // Hangul Syllables
		return true
	case r >= 0xF900 && r <= 0xFAFF: // CJK Compatibility Ideographs
		return true
	case r >= 0xFB1D && r <= 0xFB4F: // Hebrew presentation forms
		return true
	case r >= 0xFB50 && r <= 0xFDFF: // Arabic Presentation Forms-A
		return true
	case r >= 0xFE70 && r <= 0xFEFF: // Arabic Presentation Forms-B
		return true
	case r >= 0xFF66 && r <= 0xFF9D: // Halfwidth Katakana
		return true
	case r >= 0x1EE00 && r <= 0x1EEFF: // Arabic Mathematical Alphabetic Symbols
		return true
	case r >= 0x20000 && r <= 0x2A6DF: // CJK Extension B
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

const muMntOpen = `<span class="Mu-mnt">`
const muMntClose = `</span>`

func hasASCIIControl(s string) bool {
	for i := range len(s) {
		if s[i] < 0x20 {
			return true
		}
	}
	return false
}

// appendForceMonospace emits Mu-mnt cells for complex/non-ASCII runs and
// HTML-special ASCII, leaving ordinary printable ASCII bare (micron-parser-js).
func (p *Parser) appendForceMonospace(b *strings.Builder, line string) {
	if line == "" {
		return
	}
	if hasASCIIControl(line) {
		line = stripASCIIControls(line)
		if line == "" {
			return
		}
	}
	for i := 0; i < len(line); {
		c := line[i]
		if c < utf8.RuneSelf {
			if c == '&' || c == '<' || c == '>' || c < 0x20 || c >= 0x7F {
				b.WriteString(muMntOpen)
				switch c {
				case '&':
					b.WriteString("&amp;")
				case '<':
					b.WriteString("&lt;")
				case '>':
					b.WriteString("&gt;")
				default:
					if c >= 0x20 {
						b.WriteByte(c)
					}
				}
				b.WriteString(muMntClose)
			} else {
				b.WriteByte(c)
			}
			i++
			continue
		}
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
		b.WriteString(muMntOpen)
		appendHTMLText(b, line[i:end])
		b.WriteString(muMntClose)
		i = end
	}
}

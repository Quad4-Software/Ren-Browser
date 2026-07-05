// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
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

func (p *Parser) appendForceMonospace(b *strings.Builder, line string) {
	line = stripASCIIControls(line)
	for i := 0; i < len(line); {
		c := line[i]
		b.WriteString(`<span class="Mu-mnt">`)
		if c < utf8.RuneSelf {
			switch c {
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
				b.WriteByte(c)
			}
			i++
			b.WriteString(`</span>`)
			continue
		}
		r, sz := utf8.DecodeRuneInString(line[i:])
		i += sz
		switch r {
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
			b.WriteRune(r)
		}
		b.WriteString(`</span>`)
	}
}

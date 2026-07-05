// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	htmlpkg "html"
	"strconv"
	"strings"
)

func appendQuotedHTMLStyleAttr(b *strings.Builder, st Style, defaultBG string) bool {
	if !hasAnyStyle(st, defaultBG) {
		return false
	}
	b.WriteString(` style="`)
	appendStyleAttr(b, st, defaultBG)
	b.WriteByte('"')
	return true
}

func appendStyledSpanOpen(b *strings.Builder, st Style, defaultBG string) bool {
	if !hasAnyStyle(st, defaultBG) {
		return false
	}
	b.WriteString(`<span style="`)
	appendStyleAttr(b, st, defaultBG)
	b.WriteString(`">`)
	return true
}

func hasAnyStyle(st Style, defaultBG string) bool {
	if micronColorToken(st.FG) {
		return true
	}
	if st.BG != defaultBG && st.BG != "default" && micronColorToken(st.BG) {
		return true
	}
	return st.Bold || st.Underline || st.Italic
}

// appendOutput writes the styled HTML for parts directly into b. It defers
// emitting the opening <span> until a non-empty body part is encountered,
// which avoids both the intermediate per-line strings.Builder used for
// buffering body text and any "open span / no body / discard" round trips.
func (p *Parser) appendOutput(b *strings.Builder, parts []linePart, s *State) {
	var cur Style
	var styleSet bool
	var spanOpen bool

	closeSpan := func() {
		if spanOpen {
			b.WriteString(`</span>`)
			spanOpen = false
		}
		styleSet = false
	}
	ensureSpan := func() {
		if !spanOpen && styleSet {
			spanOpen = appendStyledSpanOpen(b, cur, s.DefaultBG)
		}
	}

	for i := range parts {
		pr := &parts[i]
		if pr.field != nil || pr.link != nil || pr.partial != nil {
			closeSpan()
			switch {
			case pr.field != nil:
				p.writeField(b, pr.field, s)
			case pr.link != nil:
				p.writeLink(b, pr.link, s)
			default:
				p.writePartial(b, pr.partial, s)
			}
			continue
		}
		st := pr.style
		if !styleSet || !stylesEqual(&st, &cur) {
			closeSpan()
			cur = st
			styleSet = true
		}
		if pr.html == "" && pr.text == "" {
			continue
		}
		ensureSpan()
		if pr.html != "" {
			b.WriteString(pr.html)
		} else if p.ForceMonospace {
			if s.Literal {
				p.appendForceMonospace(b, pr.text)
			} else {
				p.appendSplitAtSpaces(b, pr.text)
			}
		} else {
			appendHTMLText(b, pr.text)
		}
	}
	closeSpan()
}

func (p *Parser) writeField(b *strings.Builder, f *Field, s *State) {
	switch f.Kind {
	case FieldCheckbox:
		b.WriteString(`<label`)
		appendQuotedHTMLStyleAttr(b, f.Style, s.DefaultBG)
		b.WriteString(`><input type="checkbox" name="`)
		b.WriteString(htmlAttr(f.Name))
		b.WriteString(`" value="`)
		b.WriteString(htmlAttr(f.Value))
		b.WriteString(`"`)
		if f.Prechecked {
			b.WriteString(` checked`)
		}
		b.WriteString(`/> `)
		appendHTMLText(b, f.Label)
		b.WriteString(`</label>`)
	case FieldRadio:
		b.WriteString(`<label`)
		appendQuotedHTMLStyleAttr(b, f.Style, s.DefaultBG)
		b.WriteString(`><input type="radio" name="`)
		b.WriteString(htmlAttr(f.Name))
		b.WriteString(`" value="`)
		b.WriteString(htmlAttr(f.Value))
		b.WriteString(`"`)
		if f.Prechecked {
			b.WriteString(` checked`)
		}
		b.WriteString(`/> `)
		appendHTMLText(b, f.Label)
		b.WriteString(`</label>`)
	default:
		t := "text"
		if f.Masked {
			t = "password"
		}
		b.WriteString(`<input`)
		appendQuotedHTMLStyleAttr(b, f.Style, s.DefaultBG)
		b.WriteString(` type="`)
		b.WriteString(t)
		b.WriteString(`" name="`)
		b.WriteString(htmlAttr(f.Name))
		b.WriteString(`" value="`)
		b.WriteString(htmlAttr(f.Value))
		b.WriteString(`"`)
		if f.Width > 0 {
			b.WriteString(` size="`)
			b.WriteString(strconv.Itoa(f.Width))
			b.WriteString(`"`)
		}
		b.WriteString(`/>`)
	}
}

func (p *Parser) writeLink(b *strings.Builder, lk *Link, s *State) {
	direct := linkDirectURL(lk.URL)
	if len(lk.Fields) == 0 {
		b.WriteString(`<a class="Mu-nl" href="`)
		b.WriteString(htmlAttr(lk.URL))
		b.WriteString(`" title="`)
		b.WriteString(htmlAttr(lk.URL))
		b.WriteString(`" data-action="openNode" data-destination="`)
		b.WriteString(htmlAttr(direct))
		b.WriteString(`"`)
		appendQuotedHTMLStyleAttr(b, lk.Style, s.DefaultBG)
		b.WriteString(`>`)
		b.WriteString(lk.Label)
		b.WriteString(`</a>`)
		return
	}
	var fieldStr strings.Builder
	var reqPairs strings.Builder
	foundAll := false
	for _, f := range lk.Fields {
		if f == "*" {
			foundAll = true
			continue
		}
		if strings.Contains(f, "=") {
			if reqPairs.Len() > 0 {
				reqPairs.WriteByte('|')
			}
			reqPairs.WriteString(f)
			continue
		}
		if fieldStr.Len() > 0 {
			fieldStr.WriteByte('|')
		}
		fieldStr.WriteString(f)
	}
	if foundAll {
		fieldStr.Reset()
		fieldStr.WriteByte('*')
	}
	if reqPairs.Len() > 0 {
		q := reqPairs.String()
		if strings.Contains(direct, "`") {
			direct = direct + "|" + q
		} else {
			direct = direct + "`" + q
		}
	}
	b.WriteString(`<a class="Mu-nl" href="`)
	b.WriteString(htmlAttr(lk.URL))
	b.WriteString(`" title="`)
	b.WriteString(htmlAttr(lk.URL))
	b.WriteString(`" data-action="openNode" data-destination="`)
	b.WriteString(htmlAttr(direct))
	b.WriteString(`" data-fields="`)
	b.WriteString(htmlAttr(fieldStr.String()))
	b.WriteString(`"`)
	appendQuotedHTMLStyleAttr(b, lk.Style, s.DefaultBG)
	b.WriteString(`>`)
	b.WriteString(lk.Label)
	b.WriteString(`</a>`)
}

func (p *Parser) writePartial(b *strings.Builder, pt *Partial, s *State) {
	b.WriteString(`<div class="Mu-partial" data-partial-url="`)
	b.WriteString(htmlAttr(pt.URL))
	b.WriteString(`" data-partial-destination="`)
	b.WriteString(htmlAttr(pt.Destination))
	b.WriteString(`" data-partial-descriptor="`)
	b.WriteString(htmlAttr(pt.Descriptor))
	b.WriteString(`"`)
	if pt.PartialID != "" {
		b.WriteString(` data-partial-id="`)
		b.WriteString(htmlAttr(pt.PartialID))
		b.WriteString(`"`)
	}
	if pt.HasRefresh {
		b.WriteString(` data-partial-refresh="`)
		b.WriteString(htmlAttr(formatPartialRefresh(pt.Refresh)))
		b.WriteString(`"`)
	}
	if pt.FieldsAttr != "" {
		b.WriteString(` data-partial-fields="`)
		b.WriteString(htmlAttr(pt.FieldsAttr))
		b.WriteString(`"`)
	}
	appendQuotedHTMLStyleAttr(b, pt.Style, s.DefaultBG)
	b.WriteString(`>`)
	b.WriteString("\u29d6")
	b.WriteString(`</div>`)
}

func htmlAttr(s string) string {
	return htmlpkg.EscapeString(stripASCIIControls(s))
}

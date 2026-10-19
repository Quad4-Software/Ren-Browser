// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"regexp"
	"strings"
)

// micronStripRe mirrors micron-parser-js _MICRON_STRIP_RE: inline format
// toggles, F/B colors, and `:name anchor codes removed before slugification.
var micronStripRe = regexp.MustCompile("`[FB]T[0-9a-fA-F]{6}|`[FB][0-9a-fA-F]{3}|`:[A-Za-z0-9_-]*|`[!*_=fbacrl`<>{]")

var slugNonAlnumRe = regexp.MustCompile(`[^A-Za-z0-9]+`)

// slugifyMicron derives a same-page anchor id from heading source text,
// matching micron-parser-js slugifyMicron.
func slugifyMicron(text string) string {
	stripped := micronStripRe.ReplaceAllString(text, "")
	slug := slugNonAlnumRe.ReplaceAllString(stripped, "-")
	slug = strings.Trim(slug, "-")
	return strings.ToLower(slug)
}

const (
	headerAnchorClassMarker = `" class="micron-anchor micron-header-anchor"`
	emptyHrefAttr           = `href="#"`
	emptyTitleAttr          = ` title="#"`
)

// resolveEmptyAnchorHrefs rewrites href="#" links to the id of the first
// micron-header-anchor that follows them, matching micron-parser-js
// _resolveEmptyAnchors. The link title is updated to match.
func resolveEmptyAnchorHrefs(html string) string {
	if !strings.Contains(html, emptyHrefAttr) || !strings.Contains(html, headerAnchorClassMarker) {
		return html
	}

	type headerAnchor struct {
		pos int
		id  string
	}
	var headers []headerAnchor
	for off := 0; off < len(html); {
		idx := strings.Index(html[off:], headerAnchorClassMarker)
		if idx < 0 {
			break
		}
		pos := off + idx
		start := pos
		for start > 0 && pos-start < 128 && html[start-1] != '<' {
			start--
		}
		seg := html[start:pos]
		if _, after, ok := strings.Cut(seg, `id="`); ok {
			headers = append(headers, headerAnchor{pos: pos, id: after})
		}
		off = pos + 1
	}
	if len(headers) == 0 {
		return html
	}

	var out strings.Builder
	out.Grow(len(html) + 64)
	last := 0
	hi := 0
	for {
		idx := strings.Index(html[last:], emptyHrefAttr)
		if idx < 0 {
			break
		}
		pos := last + idx
		for hi < len(headers) && headers[hi].pos <= pos {
			hi++
		}
		if hi >= len(headers) {
			break
		}
		id := headers[hi].id
		out.WriteString(html[last:pos])
		out.WriteString(`href="#`)
		out.WriteString(id)
		out.WriteByte('"')
		pos += len(emptyHrefAttr)
		if strings.HasPrefix(html[pos:], emptyTitleAttr) {
			out.WriteString(` title="#`)
			out.WriteString(id)
			out.WriteByte('"')
			pos += len(emptyTitleAttr)
		}
		last = pos
	}
	out.WriteString(html[last:])
	return out.String()
}

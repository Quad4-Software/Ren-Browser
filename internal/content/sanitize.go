// SPDX-License-Identifier: MIT
package content

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	xhtml "golang.org/x/net/html"
)

const (
	maxSanitizedHTMLNodes   = 20000
	maxSanitizedHTMLDepth   = 256
	maxSanitizedAttrBytes   = 2048
	maxSanitizedClassBytes  = 1024
	maxSanitizedURLBytes    = 2048
	maxSanitizedOutputBytes = 12 * 1024 * 1024
)

var javascriptSchemeRe = regexp.MustCompile(`(?i)javascript:`)

var allowedSanitizedTags = map[string]struct{}{
	"a": {}, "abbr": {}, "address": {}, "article": {}, "aside": {}, "b": {},
	"blockquote": {}, "br": {}, "caption": {}, "cite": {}, "code": {}, "col": {},
	"colgroup": {}, "dd": {}, "del": {}, "details": {}, "dfn": {}, "div": {},
	"dl": {}, "dt": {}, "em": {}, "figcaption": {}, "figure": {}, "footer": {},
	"h1": {}, "h2": {}, "h3": {}, "h4": {}, "h5": {}, "h6": {}, "header": {},
	"hgroup": {}, "hr": {}, "i": {}, "img": {}, "ins": {}, "kbd": {}, "li": {},
	"main": {}, "mark": {}, "nav": {}, "ol": {}, "p": {}, "pre": {}, "q": {},
	"rp": {}, "rt": {}, "ruby": {}, "s": {}, "samp": {}, "section": {}, "small": {},
	"span": {}, "strong": {}, "sub": {}, "summary": {}, "sup": {}, "table": {},
	"tbody": {}, "td": {}, "tfoot": {}, "th": {}, "thead": {}, "time": {}, "tr": {},
	"u": {}, "ul": {}, "var": {}, "wbr": {},
}

var forbiddenSanitizedTags = map[string]struct{}{
	"applet": {}, "audio": {}, "base": {}, "button": {}, "embed": {}, "fencedframe": {},
	"form": {}, "frame": {}, "frameset": {}, "iframe": {}, "input": {}, "link": {},
	"math": {}, "meta": {}, "noscript": {}, "object": {}, "option": {}, "picture": {},
	"plaintext": {}, "portal": {}, "script": {}, "select": {}, "source": {}, "style": {},
	"svg": {}, "template": {}, "textarea": {}, "track": {}, "video": {}, "xmp": {},
}

var allowedSanitizedAttrs = map[string]struct{}{
	"abbr": {}, "align": {}, "alt": {}, "aria-describedby": {}, "aria-hidden": {},
	"aria-label": {}, "aria-labelledby": {}, "class": {}, "colspan": {}, "datetime": {},
	"dir": {}, "headers": {}, "height": {}, "id": {}, "lang": {}, "reversed": {},
	"rowspan": {}, "scope": {}, "span": {}, "start": {}, "title": {}, "type": {},
	"value": {}, "width": {},
}

func SanitizeHTML(input string) string {
	if input == "" {
		return input
	}
	if strings.IndexByte(input, '<') < 0 {
		return stripJavaScriptSchemeText(input)
	}
	if clearlyPassiveHTML(input) {
		return input
	}
	z := xhtml.NewTokenizer(strings.NewReader(input))
	var out strings.Builder
	out.Grow(len(input))
	budget := maxSanitizedHTMLNodes
	depth := 0
	skipTag := ""
	skipDepth := 0
	for budget > 0 && out.Len() < maxSanitizedOutputBytes {
		tokenType := z.Next()
		if tokenType == xhtml.ErrorToken {
			break
		}
		switch tokenType {
		case xhtml.TextToken:
			if skipDepth != 0 {
				continue
			}
			budget--
			out.WriteString(html.EscapeString(stripJavaScriptSchemeText(string(z.Text()))))
		case xhtml.StartTagToken, xhtml.SelfClosingTagToken:
			name, _ := z.TagName()
			tag := strings.ToLower(string(name))
			if skipDepth != 0 {
				if tag == skipTag {
					skipDepth++
				}
				continue
			}
			if _, forbidden := forbiddenSanitizedTags[tag]; forbidden {
				if !voidSanitizedTag(tag) && tokenType == xhtml.StartTagToken {
					skipTag = tag
					skipDepth = 1
				}
				continue
			}
			if _, allowed := allowedSanitizedTags[tag]; !allowed || depth >= maxSanitizedHTMLDepth {
				continue
			}
			budget--
			out.WriteByte('<')
			out.WriteString(tag)
			writeSanitizedAttrs(&out, z, tag)
			out.WriteByte('>')
			if voidSanitizedTag(tag) {
				continue
			}
			if tokenType == xhtml.SelfClosingTagToken {
				out.WriteString("</")
				out.WriteString(tag)
				out.WriteByte('>')
				continue
			}
			depth++
		case xhtml.EndTagToken:
			name, _ := z.TagName()
			tag := strings.ToLower(string(name))
			if skipDepth != 0 {
				if tag == skipTag {
					skipDepth--
					if skipDepth == 0 {
						skipTag = ""
					}
				}
				continue
			}
			if _, allowed := allowedSanitizedTags[tag]; !allowed || voidSanitizedTag(tag) {
				continue
			}
			out.WriteString("</")
			out.WriteString(tag)
			out.WriteByte('>')
			if depth > 0 {
				depth--
			}
		}
	}
	if out.Len() > maxSanitizedOutputBytes {
		return out.String()[:maxSanitizedOutputBytes]
	}
	return out.String()
}

func clearlyPassiveHTML(input string) bool {
	if strings.ContainsAny(input, "&=") {
		return false
	}
	lower := strings.ToLower(input)
	if strings.Contains(lower, "javascript:") {
		return false
	}
	depth := 0
	for i := 0; i < len(lower); i++ {
		c := lower[i]
		if c < 0x20 || c == 0x7f {
			return false
		}
		if c != '<' {
			continue
		}
		j := i + 1
		endTag := false
		if j < len(lower) && lower[j] == '/' {
			endTag = true
			j++
		}
		start := j
		for j < len(lower) && ((lower[j] >= 'a' && lower[j] <= 'z') || (lower[j] >= '0' && lower[j] <= '9')) {
			j++
		}
		if start == j {
			continue
		}
		name := lower[start:j]
		if _, forbidden := forbiddenSanitizedTags[name]; forbidden {
			return false
		}
		if endTag {
			if depth > 0 {
				depth--
			}
			continue
		}
		if !voidSanitizedTag(name) {
			depth++
			if depth > maxSanitizedHTMLDepth {
				return false
			}
		}
	}
	return true
}

func writeSanitizedAttrs(out *strings.Builder, z *xhtml.Tokenizer, tag string) {
	count := 0
	for {
		key, val, more := z.TagAttr()
		if len(key) == 0 {
			break
		}
		name := strings.ToLower(string(key))
		value := string(val)
		if count < 64 && validSanitizedAttr(tag, name, value) {
			out.WriteByte(' ')
			out.WriteString(name)
			out.WriteString(`="`)
			out.WriteString(html.EscapeString(value))
			out.WriteByte('"')
		}
		count++
		if !more {
			break
		}
	}
}

func validSanitizedAttr(tag, name, value string) bool {
	if name == "" || strings.HasPrefix(name, "on") || !utf8.ValidString(value) || len(value) > maxSanitizedAttrBytes {
		return false
	}
	if strings.HasPrefix(name, "data-") {
		return true
	}
	if name == "class" && len(value) > maxSanitizedClassBytes {
		return false
	}
	switch name {
	case "href":
		return tag == "a" && safeSanitizedLinkURL(value)
	case "src":
		return tag == "img" && safeSanitizedImageURL(value)
	case "srcset", "srcdoc", "style", "formaction", "action", "poster", "background", "xlink:href":
		return false
	default:
		_, ok := allowedSanitizedAttrs[name]
		return ok
	}
}

func stripJavaScriptSchemeText(input string) string {
	for range 3 {
		out := javascriptSchemeRe.ReplaceAllString(input, "")
		if out == input {
			return input
		}
		input = out
	}
	return input
}

func voidSanitizedTag(tag string) bool {
	switch tag {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr":
		return true
	}
	return false
}

func safeSanitizedLinkURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > maxSanitizedURLBytes || hasURLControlByte(raw) || strings.HasPrefix(raw, "//") {
		return false
	}
	colon := strings.IndexByte(raw, ':')
	if colon < 0 {
		return true
	}
	scheme := strings.ToLower(raw[:colon])
	if len(scheme) == 32 && isLowerHex(scheme) {
		return true
	}
	return scheme == "rns"
}

func safeSanitizedImageURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if len(raw) == 0 || len(raw) > maxSanitizedURLBytes || hasURLControlByte(raw) {
		return false
	}
	lower := strings.ToLower(raw)
	for _, mime := range []string{"png", "jpeg", "jpg", "gif", "webp", "bmp", "tiff"} {
		if strings.HasPrefix(lower, "data:image/"+mime+";base64,") {
			return true
		}
	}
	return false
}

func hasURLControlByte(raw string) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] <= 0x20 || raw[i] == 0x7f {
			return true
		}
	}
	return false
}

func isLowerHex(raw string) bool {
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

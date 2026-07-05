package content

import (
	"regexp"
	"strings"
)

var (
	scriptBlockRe = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	styleBlockRe  = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	onAttrRe      = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	iframeRe      = regexp.MustCompile(`(?is)<iframe\b[^>]*>.*?</iframe>`)
	objectRe      = regexp.MustCompile(`(?is)<object\b[^>]*>.*?</object>`)
	embedRe       = regexp.MustCompile(`(?is)<embed\b[^>]*/?>`)
)

func SanitizeHTML(input string) string {
	if input == "" || !needsHTMLSanitize(input) {
		return input
	}
	out := input
	out = scriptBlockRe.ReplaceAllString(out, "")
	out = styleBlockRe.ReplaceAllString(out, "")
	out = iframeRe.ReplaceAllString(out, "")
	out = objectRe.ReplaceAllString(out, "")
	out = embedRe.ReplaceAllString(out, "")
	out = onAttrRe.ReplaceAllString(out, "")
	out = strings.ReplaceAll(out, "javascript:", "")
	return out
}

func needsHTMLSanitize(s string) bool {
	n := len(s)
	for i := 0; i < n; i++ {
		switch s[i] {
		case 'j', 'J':
			if i+10 < n && asciiEqualFold(s[i:i+11], "javascript:") {
				return true
			}
		case '<':
			if htmlTagAt(s, i, "script") ||
				htmlTagAt(s, i, "style") ||
				htmlTagAt(s, i, "iframe") ||
				htmlTagAt(s, i, "object") ||
				htmlTagAt(s, i, "embed") {
				return true
			}
		case ' ', '\t', '\n', '\r':
			if i+3 < n && asciiEqualFold(s[i+1:i+3], "on") {
				c := s[i+3]
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
					return true
				}
			}
		}
	}
	return false
}

func htmlTagAt(s string, i int, tag string) bool {
	if i+1+len(tag) > len(s) || s[i] != '<' {
		return false
	}
	for j := 0; j < len(tag); j++ {
		if !asciiEqualFoldByte(s[i+1+j], tag[j]) {
			return false
		}
	}
	if i+1+len(tag) == len(s) {
		return true
	}
	switch s[i+1+len(tag)] {
	case '>', ' ', '\t', '\n', '\r', '/':
		return true
	default:
		return false
	}
}

func asciiEqualFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !asciiEqualFoldByte(a[i], b[i]) {
			return false
		}
	}
	return true
}

func asciiEqualFoldByte(a, b byte) bool {
	if a >= 'A' && a <= 'Z' {
		a += 'a' - 'A'
	}
	if b >= 'A' && b <= 'Z' {
		b += 'a' - 'A'
	}
	return a == b
}

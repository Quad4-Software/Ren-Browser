// SPDX-License-Identifier: MIT
package nomadnet

import (
	"strings"
)

type ContentKind string

const (
	KindMicron    ContentKind = "micron"
	KindHTML      ContentKind = "html"
	KindMarkdown  ContentKind = "markdown"
	KindPlaintext ContentKind = "plaintext"
	KindBinary    ContentKind = "binary"
)

func DetectContentType(path string, body []byte) string {
	path = strings.ToLower(path)
	switch {
	case strings.HasSuffix(path, ".mu"):
		return string(KindMicron)
	case strings.HasSuffix(path, ".html"), strings.HasSuffix(path, ".htm"):
		return string(KindHTML)
	case strings.HasSuffix(path, ".md"), strings.HasSuffix(path, ".markdown"):
		return string(KindMarkdown)
	case strings.HasSuffix(path, ".txt"):
		return string(KindPlaintext)
	}

	sample := body
	if len(sample) > 512 {
		sample = sample[:512]
	}
	if len(sample) == 0 {
		return string(KindPlaintext)
	}
	trim := trimSpaceBytes(sample)
	switch {
	case hasPrefixFold(trim, "<!doctype") || hasPrefixFold(trim, "<html"):
		return string(KindHTML)
	case bytesContains(trim, "`>") || bytesContains(trim, "`>>"):
		return string(KindMicron)
	case hasPrefix(trim, []byte("# ")) || bytesContains(trim, "\n## "):
		return string(KindMarkdown)
	default:
		return string(KindPlaintext)
	}
}

func trimSpaceBytes(b []byte) []byte {
	start, end := 0, len(b)
	for start < end && isSpace(b[start]) {
		start++
	}
	for end > start && isSpace(b[end-1]) {
		end--
	}
	return b[start:end]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func hasPrefix(b, prefix []byte) bool {
	if len(b) < len(prefix) {
		return false
	}
	for i := range prefix {
		if b[i] != prefix[i] {
			return false
		}
	}
	return true
}

func hasPrefixFold(b []byte, prefix string) bool {
	if len(b) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		a, c := b[i], prefix[i]
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if a != c {
			return false
		}
	}
	return true
}

func bytesContains(b []byte, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	if len(b) < len(sub) {
		return false
	}
	for i := 0; i <= len(b)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if b[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// SPDX-License-Identifier: MIT
package nomadnet

import (
	"bytes"
)

type ContentKind string

const (
	KindMicron    ContentKind = "micron"
	KindHTML      ContentKind = "html"
	KindMarkdown  ContentKind = "markdown"
	KindPlaintext ContentKind = "plaintext"
	KindPDF       ContentKind = "pdf"
	KindEPUB      ContentKind = "epub"
	KindBinary    ContentKind = "binary"
)

func DetectContentType(path string, body []byte) string {
	if kind, ok := contentKindByPath(path); ok {
		return kind
	}
	if kind := probeDocumentKind(body); kind != "" {
		return kind
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
	case bytes.Contains(trim, []byte("`>")) || bytes.Contains(trim, []byte("`>>")):
		return string(KindMicron)
	case hasPrefix(trim, []byte("# ")) || bytes.Contains(trim, []byte("\n## ")):
		return string(KindMarkdown)
	default:
		return string(KindPlaintext)
	}
}

func contentKindByPath(path string) (string, bool) {
	switch {
	case pathEndsWithFold(path, ".mu"):
		return string(KindMicron), true
	case pathEndsWithFold(path, ".html"), pathEndsWithFold(path, ".htm"):
		return string(KindHTML), true
	case pathEndsWithFold(path, ".md"), pathEndsWithFold(path, ".markdown"):
		return string(KindMarkdown), true
	case pathEndsWithFold(path, ".txt"):
		return string(KindPlaintext), true
	case pathEndsWithFold(path, ".pdf"):
		return string(KindPDF), true
	case pathEndsWithFold(path, ".epub"):
		return string(KindEPUB), true
	default:
		return "", false
	}
}

func IsDocumentKind(kind string) bool {
	switch ContentKind(kind) {
	case KindPDF, KindEPUB:
		return true
	default:
		return false
	}
}

func probeDocumentKind(body []byte) string {
	if len(body) >= 5 && bytes.Equal(body[:5], []byte("%PDF-")) {
		return string(KindPDF)
	}
	if len(body) >= 30 && bytes.Equal(body[:4], []byte("PK\x03\x04")) {
		limit := body
		if len(limit) > 512 {
			limit = limit[:512]
		}
		if bytes.Contains(limit, []byte("application/epub+zip")) {
			return string(KindEPUB)
		}
	}
	return ""
}

func pathEndsWithFold(path, suffix string) bool {
	if len(path) < len(suffix) {
		return false
	}
	part := path[len(path)-len(suffix):]
	for i := 0; i < len(suffix); i++ {
		a, b := part[i], suffix[i]
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
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
	return bytes.Equal(b[:len(prefix)], prefix)
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

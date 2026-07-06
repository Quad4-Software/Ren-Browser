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
	KindBinary    ContentKind = "binary"
)

func DetectContentType(path string, body []byte) string {
	if kind, ok := contentKindByPath(path); ok {
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
	default:
		return "", false
	}
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

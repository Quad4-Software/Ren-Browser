// SPDX-License-Identifier: MIT
package content_test

import (
	"strings"
	"testing"

	"renbrowser/internal/content"
)

func FuzzSanitizeHTML(f *testing.F) {
	f.Add(`<p>hello</p>`)
	f.Add(`<script>alert(1)</script><b>ok</b>`)
	f.Add(`<a href="javascript:alert(1)">x</a>`)
	f.Add(`<img onerror="alert(1)" src="/x">`)

	f.Fuzz(func(t *testing.T, input string) {
		out := content.SanitizeHTML(input)
		if strings.Contains(strings.ToLower(out), "<script") {
			t.Fatalf("script tag survived: %q", out)
		}
		if strings.Contains(strings.ToLower(out), "javascript:") {
			t.Fatalf("javascript: scheme survived: %q", out)
		}
	})
}

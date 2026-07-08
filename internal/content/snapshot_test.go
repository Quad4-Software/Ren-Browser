// SPDX-License-Identifier: MIT
package content_test

import (
	"strings"
	"testing"

	"renbrowser/internal/content"
)

func TestSanitizeHTMLSnapshotStable(t *testing.T) {
	in := `<div class="page"><p>Hello</p><script>alert(1)</script><a href="/x" onclick="evil()">link</a></div>`
	a := content.SanitizeHTML(in)
	b := content.SanitizeHTML(in)
	if a != b {
		t.Fatalf("sanitize output not stable")
	}
	if strings.Contains(strings.ToLower(a), "<script") {
		t.Fatalf("script survived: %s", a)
	}
	if strings.Contains(strings.ToLower(a), "onclick") {
		t.Fatalf("onclick survived: %s", a)
	}
	if !strings.Contains(a, "Hello") || !strings.Contains(a, "link") {
		t.Fatalf("expected content missing: %s", a)
	}
}

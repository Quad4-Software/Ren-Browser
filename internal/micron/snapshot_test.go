// SPDX-License-Identifier: MIT
package micron_test

import (
	"strings"
	"testing"

	"renbrowser/internal/micron"
)

func TestMicronToHTMLDarkSnapshot(t *testing.T) {
	src := strings.Join([]string{
		"`>Welcome",
		"Hello mesh.",
		"`[Home`/page/index.mu]",
		"`!",
		"aside note",
	}, "\n")
	html := micron.ToHTMLDark(src)
	plain := plainText(html)
	want := []string{
		"Welcome",
		"Hello mesh.",
		"aside note",
	}
	for _, fragment := range want {
		if !strings.Contains(plain, fragment) {
			t.Fatalf("snapshot missing %q in:\n%s", fragment, html)
		}
	}
	if !strings.Contains(html, "/page/index.mu") {
		t.Fatalf("snapshot missing link href in:\n%s", html)
	}
	if !strings.Contains(html, `class="Mu-mnt"`) {
		t.Fatalf("expected Mu-mnt cells in:\n%s", html)
	}
	if strings.Contains(strings.ToLower(html), "<script") {
		t.Fatalf("unexpected script in micron html: %s", html)
	}
}

func TestMicronToHTMLStableStructure(t *testing.T) {
	src := "`>Title\nBody"
	a := micron.ToHTMLDark(src)
	b := micron.ToHTMLDark(src)
	if a != b {
		t.Fatalf("micron render not stable\n---\n%s\n---\n%s", a, b)
	}
}

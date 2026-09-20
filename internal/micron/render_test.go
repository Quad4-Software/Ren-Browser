// SPDX-License-Identifier: MIT
package micron_test

import (
	"regexp"
	"strings"
	"testing"

	"renbrowser/internal/micron"
)

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func plainText(html string) string {
	return htmlTagRe.ReplaceAllString(html, "")
}

func TestMicronHeadingAndLink(t *testing.T) {
	src := "`>Welcome\n`[Home`/page/index.mu]"
	html := micron.ToHTMLDark(src)
	plain := plainText(html)
	if !strings.Contains(plain, "Welcome") {
		t.Fatalf("missing heading text: %s", html)
	}
	if !strings.Contains(html, `/page/index.mu`) {
		t.Fatalf("missing link href: %s", html)
	}
	if !strings.Contains(html, "Mu-mnt") {
		t.Fatalf("expected force-monospace markup (Mu-mnt or Mu-mnt-group): %s", html)
	}
}

func TestRenderDarkAlwaysForceMonospace(t *testing.T) {
	html, _, _ := micron.RenderDark("|=|\n|A|")
	if !strings.Contains(html, "Mu-mnt") {
		t.Fatalf("ASCII art must use monospace markup: %s", html)
	}
	if !strings.Contains(plainText(html), "|A|") {
		t.Fatalf("ASCII art text lost: %s", html)
	}
}

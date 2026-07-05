// SPDX-License-Identifier: MIT
package micron_test

import (
	"strings"
	"testing"

	"renbrowser/internal/micron"
)

func TestMicronHeadingAndLink(t *testing.T) {
	src := "`>Welcome\n`[Home`/page/index.mu]"
	html := micron.ToHTMLDark(src)
	if !strings.Contains(html, "Welcome") {
		t.Fatalf("missing heading text: %s", html)
	}
	if !strings.Contains(html, `/page/index.mu`) {
		t.Fatalf("missing link href: %s", html)
	}
}

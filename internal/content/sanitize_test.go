// SPDX-License-Identifier: MIT
package content_test

import (
	"strings"
	"testing"

	"renbrowser/internal/content"
)

func TestSanitizeHTMLRemovesScript(t *testing.T) {
	in := `<div>ok<script>alert(1)</script></div>`
	out := content.SanitizeHTML(in)
	if strings.Contains(out, "<script") {
		t.Fatalf("script not removed: %s", out)
	}
	if !strings.Contains(out, "ok") {
		t.Fatalf("content stripped: %s", out)
	}
}

func TestSanitizeHTMLRemovesPartialScriptTag(t *testing.T) {
	for _, in := range []string{"<sCript0", "<SCript", "<SCRIPT>", "<SCri<SCript>pt"} {
		out := content.SanitizeHTML(in)
		if strings.Contains(strings.ToLower(out), "<script") {
			t.Fatalf("partial script tag survived for %q: %q", in, out)
		}
	}
}

func TestSanitizeHTMLRemovesOnClick(t *testing.T) {
	in := `<a href="/x" onclick="evil()">link</a>`
	out := content.SanitizeHTML(in)
	if strings.Contains(strings.ToLower(out), "onclick") {
		t.Fatalf("onclick not removed: %s", out)
	}
}

func TestSanitizeHTMLCleanFastPath(t *testing.T) {
	in := "<html><body><p>hello</p></body></html>"
	out := content.SanitizeHTML(in)
	if out != in {
		t.Fatalf("clean html changed: %q", out)
	}
}

func TestSanitizeHTMLRemovesMetaRefresh(t *testing.T) {
	in := `<html><head><meta http-equiv="refresh" content="0;url=https://example.com"></head><body>ok</body></html>`
	out := content.SanitizeHTML(in)
	if strings.Contains(strings.ToLower(out), "refresh") {
		t.Fatalf("meta refresh not removed: %s", out)
	}
	if !strings.Contains(out, "ok") {
		t.Fatalf("content stripped: %s", out)
	}
}

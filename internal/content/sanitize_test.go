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
	for _, in := range []string{"<SCript", "<SCRIPT>", "<SCri<SCript>pt"} {
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

func TestSanitizeHTMLRemovesNestedJavascriptScheme(t *testing.T) {
	in := "jAvAsjAvAsCript:Cript:"
	out := content.SanitizeHTML(in)
	if strings.Contains(strings.ToLower(out), "javascript:") {
		t.Fatalf("nested javascript: scheme survived: %q", out)
	}
}

func TestSanitizeHTMLKeepsCleanBodyContent(t *testing.T) {
	in := "<html><body><p>hello</p></body></html>"
	out := content.SanitizeHTML(in)
	if out != in {
		t.Fatalf("clean passive html changed: %q", out)
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

func TestSanitizeHTMLBlocksEncodedAndNestedAttackMarkup(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"encoded scheme", `<a href="java&#x73;cript:alert(1)">x</a>`},
		{"whitespace scheme", `<a href="java&#9;script:alert(1)">x</a>`},
		{"srcdoc", `<iframe srcdoc="<script>alert(1)</script>"></iframe>`},
		{"svg namespace", `<svg><a href="javascript:alert(1)">x</a></svg>`},
		{"split script", `<sc<script>ript>alert(1)</script>`},
		{"external image", `<img src="https://example.invalid/pixel">`},
		{"image srcset", `<img srcset="https://example.invalid/pixel 1x">`},
		{"inline overlay style", `<p style="position:fixed;inset:0;z-index:99999">x</p>`},
	}
	for _, tc := range tests {
		out := content.SanitizeHTML(tc.in)
		lower := strings.ToLower(out)
		for _, bad := range []string{"javascript:", "<script", "srcdoc", "<iframe", "<svg", "https://example.invalid", "position:fixed"} {
			if strings.Contains(lower, bad) {
				t.Fatalf("%s left %q in %q", tc.name, bad, out)
			}
		}
	}
}

func TestSanitizeHTMLAllowsMeshLinksAndDataImages(t *testing.T) {
	in := `<a href="/page/index.mu">relative</a><a href="abb3ebcd03cb2388a838e70c001291f9:/page/index.mu">mesh</a><img src="data:image/png;base64,iVBORw0KGgo=" alt="ok">`
	out := content.SanitizeHTML(in)
	for _, want := range []string{`href="/page/index.mu"`, `href="abb3ebcd03cb2388a838e70c001291f9:/page/index.mu"`, `src="data:image/png;base64,iVBORw0KGgo="`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %q", want, out)
		}
	}
}

func TestSanitizeHTMLCapsExtremeDepth(t *testing.T) {
	var b strings.Builder
	for range 600 {
		b.WriteString("<div>")
	}
	b.WriteString("deep")
	for range 600 {
		b.WriteString("</div>")
	}
	out := content.SanitizeHTML(b.String())
	if len(out) == 0 {
		t.Fatal("expected bounded partial output")
	}
	if strings.Count(out, "<div>") > 260 {
		t.Fatalf("depth not capped: %d divs", strings.Count(out, "<div>"))
	}
}

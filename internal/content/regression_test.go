// SPDX-License-Identifier: MIT
package content_test

import (
	"testing"

	"renbrowser/internal/content"
	"renbrowser/internal/plugins/builtin"
)

func TestRenderUppercaseMicronExtension(t *testing.T) {
	builtin.RegisterRenderers(content.RendererRegistry())
	out := content.Render("/page/index.MU", []byte("`>Title\nline"), "")
	if out.Kind != "micron" {
		t.Fatalf("kind=%q", out.Kind)
	}
	if out.HTML == "" {
		t.Fatal("expected html output")
	}
}

func TestSanitizeHTMLFastPathCleanDocument(t *testing.T) {
	const clean = "<html><body><p>safe</p></body></html>"
	out := content.SanitizeHTML(clean)
	if out != clean {
		t.Fatalf("sanitizer mutated clean html: %q", out)
	}
}

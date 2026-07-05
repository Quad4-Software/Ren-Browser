// SPDX-License-Identifier: MIT
package content_test

import (
	"os"
	"strings"
	"testing"

	"renbrowser/internal/content"
	"renbrowser/internal/plugins/builtin"
)

func TestMain(m *testing.M) {
	builtin.RegisterRenderers(content.RendererRegistry())
	os.Exit(m.Run())
}

func TestRenderMicron(t *testing.T) {
	out := content.Render("/page/index.mu", []byte("`>Title\nplain"), "")
	if out.Kind != "micron" {
		t.Fatalf("kind = %q", out.Kind)
	}
	if !strings.Contains(out.HTML, "Title") {
		t.Fatalf("html = %s", out.HTML)
	}
}

func TestRenderHTMLPassthrough(t *testing.T) {
	raw := "<html><body>ok</body></html>"
	out := content.Render("/page/x.html", []byte(raw), "")
	if out.HTML != raw {
		t.Fatalf("html passthrough failed")
	}
}

func TestRenderPlaintextEscaped(t *testing.T) {
	out := content.Render("/page/x.txt", []byte("<script>"), "")
	if strings.Contains(out.HTML, "<script>") {
		t.Fatalf("plaintext not escaped: %s", out.HTML)
	}
}

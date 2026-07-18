// SPDX-License-Identifier: MIT
package security_test

import (
	"runtime"
	"strings"
	"testing"

	"renbrowser/internal/content"
	"renbrowser/internal/limits"
	"renbrowser/internal/micron"
	"renbrowser/internal/plugins"
	"renbrowser/internal/plugins/builtin"
)

func TestMicronForceMonospaceAmplification(t *testing.T) {
	// RenderDark always enables ForceMonospace (per-rune Mu-mnt spans).
	const n = 64 * 1024
	src := strings.Repeat("A", n)
	html, _, _ := micron.RenderDark(src)
	ratio := float64(len(html)) / float64(len(src))
	t.Logf("input=%d html=%d ratio=%.1fx", len(src), len(html), ratio)
	if ratio < 20 {
		t.Fatalf("expected ForceMonospace amplification >=20x, got %.1fx", ratio)
	}
	// Extrapolate to default page cap.
	pageCap := limits.DefaultMaxPageBytes
	projected := int64(float64(pageCap) * ratio)
	t.Logf("projected HTML at %d page cap: %d bytes (%.1f MiB)", pageCap, projected, float64(projected)/(1024*1024))
	if projected < 100*1024*1024 {
		t.Fatalf("expected projected HTML over 100MiB at page cap, got %d", projected)
	}
}

func TestMicronLeadingAngleRecursion(t *testing.T) {
	// parseLineInto recurses once per leading '<'. Deep stacks can panic.
	depths := []int{1000, 10000, 50000}
	for _, depth := range depths {
		depth := depth
		t.Run(itoa(depth), func(t *testing.T) {
			done := make(chan any, 1)
			go func() {
				defer func() { done <- recover() }()
				src := strings.Repeat("<", depth) + "x"
				_ = micron.ToHTMLDark(src)
				done <- nil
			}()
			if err := <-done; err != nil {
				t.Logf("depth=%d recovered panic: %v", depth, err)
				if depth < 10000 {
					t.Fatalf("unexpected panic at modest depth %d: %v", depth, err)
				}
				return
			}
			t.Logf("depth=%d completed without panic", depth)
		})
	}
}

func TestMicronBuiltinRendererSkipsSanitizeHTML(t *testing.T) {
	body := []byte("hello <script>alert(1)</script>")
	reg := plugins.NewRegistry()
	builtin.RegisterRenderers(reg)
	renderer, ok := reg.BestRenderer("/page/index.mu", body, "micron")
	if !ok || renderer.ID() != "builtin.micron" {
		t.Fatal("builtin.micron renderer missing")
	}
	out, err := renderer.Render("/page/index.mu", body, "aabbccddeeff00112233445566778899")
	if err != nil {
		t.Fatal(err)
	}
	// Go micron path does not call SanitizeHTML. ForceMonospace escapes per rune,
	// so "&lt;script" is not contiguous across Mu-mnt spans.
	if strings.Contains(out.HTML, "<script>") {
		t.Fatalf("micron HTML contains raw script tag: %s", truncate(out.HTML, 400))
	}
	if !strings.Contains(out.HTML, "&lt;") {
		t.Fatalf("expected escaped '<' entities, got: %s", truncate(out.HTML, 400))
	}
	_ = content.SanitizeHTML
	t.Logf("micron escaped '<' via parser; SanitizeHTML skipped by design (html=%d bytes)", len(out.HTML))
}

func TestMicronHeadingDepthUnbounded(t *testing.T) {
	src := strings.Repeat(">", 5000) + "Title"
	html := micron.ToHTMLDark(src)
	// ForceMonospace splits "Title" into per-rune spans.
	if !strings.Contains(html, `>T</span>`) || !strings.Contains(html, `>e</span>`) {
		t.Fatalf("missing heading glyphs in html (%d bytes): %s", len(html), truncate(html, 400))
	}
	// Depth is capped at 16: margin-left = (16-1)*2*0.6 = 18.0em
	if !strings.Contains(html, "margin-left:18.0em") {
		t.Fatalf("expected capped margin-left:18.0em, got: %s", truncate(html, 400))
	}
	if strings.Contains(html, "margin-left:5998.8em") {
		t.Fatal("uncapped heading depth still present")
	}
	t.Logf("heading depth 5000 capped -> html %d bytes with margin-left:18.0em", len(html))
}

func TestMicronPageCapStillAllowsHugeHTML(t *testing.T) {
	// Even a small fraction of MaxPageBytes blows up after ForceMonospace.
	budget := 256 * 1024
	src := strings.Repeat("W", budget)
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	html, _, _ := micron.RenderDark(src)
	runtime.ReadMemStats(&after)
	delta := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	t.Logf("src=%d html=%d heapDelta≈%d", len(src), len(html), delta)
	if len(html) < budget*20 {
		t.Fatalf("html %d too small for ForceMonospace on %d input", len(html), budget)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

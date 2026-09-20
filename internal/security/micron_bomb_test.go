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
	// RenderDark always enables ForceMonospace. micron-parser-go v1.2.0
	// groups plain printable ASCII runs into a single Mu-mnt-group span, so
	// ASCII input stays near 1x. HTML-significant bytes and grapheme
	// clusters still get per-rune Mu-mnt cells, which is the residual
	// worst case and must stay bounded.
	const n = 64 * 1024
	for _, tc := range []struct {
		name    string
		r       string
		maxMult float64
	}{
		{"ascii", "A", 2.0},
		{"html-significant", "&", 40.0},
		{"emoji", "\U0001F642", 12.0},
	} {
		src := strings.Repeat(tc.r, n/len(tc.r))
		html, _, _ := micron.RenderDark(src)
		ratio := float64(len(html)) / float64(len(src))
		t.Logf("%s input=%d html=%d ratio=%.1fx", tc.name, len(src), len(html), ratio)
		if ratio > tc.maxMult {
			t.Fatalf("%s amplification %.1fx exceeds bound %.1fx", tc.name, ratio, tc.maxMult)
		}
	}
	// Extrapolate the residual worst case to the default page cap.
	pageCap := limits.DefaultMaxPageBytes
	projected := int64(float64(pageCap) * 33)
	t.Logf("projected worst-case HTML at %d page cap: %d bytes (%.1f MiB)", pageCap, projected, float64(projected)/(1024*1024))
}

func TestMicronLeadingAngleRecursion(t *testing.T) {
	// parseLineInto recurses once per leading '<'. Deep stacks can panic.
	depths := []int{1000, 10000, 50000}
	for _, depth := range depths {
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
	// v1.2.0 groups plain ASCII into Mu-mnt-group spans; assert the heading
	// text survived whatever wrapping the parser chose.
	if !strings.Contains(html, "Title") {
		t.Fatalf("missing heading text in html (%d bytes): %s", len(html), truncate(html, 400))
	}
	// Depth is capped at 16: indent = (16-1)*2*0.6 = 18.0em.
	// micron-parser-go v1.0.7+ emits margin-inline-start (older builds used margin-left).
	capped := strings.Contains(html, "margin-inline-start:18.0em") ||
		strings.Contains(html, "margin-left:18.0em")
	if !capped {
		t.Fatalf("expected capped margin-inline-start/margin-left:18.0em, got: %s", truncate(html, 400))
	}
	if strings.Contains(html, "margin-inline-start:5998.8em") ||
		strings.Contains(html, "margin-left:5998.8em") {
		t.Fatal("uncapped heading depth still present")
	}
	t.Logf("heading depth 5000 capped -> html %d bytes with 18.0em indent", len(html))
}

func TestMicronPageCapHTMLAmplificationBounds(t *testing.T) {
	// A page-capped input used to blow up ~30x through per-rune Mu-mnt
	// spans. v1.2.0 groups ASCII so plain text stays near 1x, while
	// HTML-significant bytes remain the bounded worst case.
	budget := 256 * 1024
	for _, tc := range []struct {
		name    string
		r       string
		maxMult int64
	}{
		{"ascii", "W", 2},
		{"html-significant", "&", 40},
	} {
		src := strings.Repeat(tc.r, budget)
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		html, _, _ := micron.RenderDark(src)
		runtime.ReadMemStats(&after)
		delta := int64(after.HeapAlloc) - int64(before.HeapAlloc)
		t.Logf("%s src=%d html=%d heapDelta≈%d", tc.name, len(src), len(html), delta)
		if int64(len(html)) > int64(budget)*tc.maxMult {
			t.Fatalf("%s html %d exceeds %dx bound on %d input", tc.name, len(html), tc.maxMult, budget)
		}
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

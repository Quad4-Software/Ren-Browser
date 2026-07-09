// SPDX-License-Identifier: MIT
package regression_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"regexp"
	"strings"
	"testing"
	"time"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/identity"
	"quad4/reticulum-go/pkg/interfaces"
	"quad4/reticulum-go/pkg/transport"

	"renbrowser/internal/content"
	"renbrowser/internal/micron"
	"renbrowser/internal/nomadnet"
	"renbrowser/internal/plugins/builtin"
)

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func plainText(html string) string {
	return htmlTagRe.ReplaceAllString(html, "")
}

func testTransport(t *testing.T) *transport.Transport {
	t.Helper()
	tr := transport.NewTransport(&common.ReticulumConfig{EnableTransport: true})
	t.Cleanup(func() { _ = tr.Close() })
	ident, err := identity.NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	tr.SetIdentity(ident)
	return tr
}

func registerPath(t *testing.T, tr *transport.Transport, dest []byte, name string) {
	t.Helper()
	udp, err := interfaces.NewUDPInterface(name, "127.0.0.1:0", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tr.RegisterInterface(name, udp); err != nil {
		t.Fatal(err)
	}
	tr.UpdatePath(dest, bytes.Repeat([]byte{0x11}, 16), name, 1)
	if !tr.HasPath(dest) {
		t.Fatal("expected path")
	}
}

func TestRegressionRefreshPathForRetryExpiresCachedRoute(t *testing.T) {
	tr := testTransport(t)
	browser := nomadnet.NewBrowser(tr, nomadnet.NewAnnounceHandler())
	dest := bytes.Repeat([]byte{0xab}, 16)
	registerPath(t, tr, dest, "wake-reg")

	browser.RefreshPathForRetry(hex.EncodeToString(dest))
	if tr.HasPath(dest) {
		t.Fatal("RefreshPathForRetry must expire cached path after suspend-style invalidation")
	}

	res := browser.PrepareForWake()
	if res.DroppedLinks < 0 || res.ExpiredPaths < 0 {
		t.Fatalf("PrepareForWake counts: %+v", res)
	}
}

func TestRegressionShouldRefreshPathAfterSuspendErrors(t *testing.T) {
	cases := map[string]bool{
		"no path to node: path discovery timed out": true,
		"link establish timeout":                    true,
		"connection lost":                           true,
		"page not found":                            false,
		"":                                          false,
	}
	for msg, want := range cases {
		if got := nomadnet.ShouldRefreshPath(msg); got != want {
			t.Fatalf("ShouldRefreshPath(%q)=%v want %v", msg, got, want)
		}
	}
}

func TestRegressionFetchMissingRouteRespectsContext(t *testing.T) {
	tr := testTransport(t)
	browser := nomadnet.NewBrowser(tr, nomadnet.NewAnnounceHandler())
	dest := bytes.Repeat([]byte{0xcd}, 16)

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	res := browser.Fetch(ctx, hex.EncodeToString(dest), "/page/index.mu", nomadnet.RequestData{})
	elapsed := time.Since(start)

	if res.Error == "" {
		t.Fatal("expected missing-route failure")
	}
	if !nomadnet.ShouldRefreshPath(res.Error) {
		t.Fatalf("missing-route error should trigger path refresh: %q", res.Error)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("missing route took too long: %v (error=%q)", elapsed, res.Error)
	}
}

func TestRegressionMicronForceMonospaceASCII(t *testing.T) {
	builtin.RegisterRenderers(content.RendererRegistry())

	art := strings.Join([]string{
		"+-----+",
		"| ASCII |",
		"+-----+",
	}, "\n")

	html, _, _ := micron.RenderDark(art)
	if !strings.Contains(html, `class="Mu-mnt"`) {
		t.Fatalf("RenderDark must emit Mu-mnt for ASCII:\n%s", html)
	}
	if !strings.Contains(plainText(html), "ASCII") {
		t.Fatalf("missing ASCII text in:\n%s", html)
	}

	out := content.Render("/page/art.mu", []byte(art), "")
	if out.Kind != "micron" {
		t.Fatalf("kind=%q", out.Kind)
	}
	if !strings.Contains(out.HTML, `class="Mu-mnt"`) {
		t.Fatalf("content.Render must emit Mu-mnt:\n%s", out.HTML)
	}
	if !strings.Contains(plainText(out.HTML), "ASCII") {
		t.Fatalf("content.Render lost ASCII text:\n%s", out.HTML)
	}
}

func TestRegressionMicronForceMonospaceAlwaysOn(t *testing.T) {
	// Historical bug: ForceMonospace was gated by MicronPreserveLayout (default
	// false), so Auto→Go HTML dropped Mu-mnt cells and broke ASCII art.
	html, _, _ := micron.RenderDark("|=|")
	if !strings.Contains(html, `class="Mu-mnt"`) {
		t.Fatalf("expected Mu-mnt cells:\n%s", html)
	}
	cells := strings.Count(html, `class="Mu-mnt"`)
	if cells < 3 {
		t.Fatalf("expected per-character cells, got %d in:\n%s", cells, html)
	}
}

//go:build interop

// SPDX-License-Identifier: MIT

package interop_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"renbrowser/internal/app"
	"renbrowser/internal/nomadnet"
	"renbrowser/internal/rns"
)

// mediaTestNode returns the node hash under test. The test is opt-in: it
// only runs when RENBROWSER_MEDIA_TEST_NODE names a NomadNet node that
// serves /page/index.mu plus /media/demo.webp and /media/test.png, since
// public nodes cannot be relied on to host known media.
func mediaTestNode(t *testing.T) string {
	t.Helper()
	node := os.Getenv("RENBROWSER_MEDIA_TEST_NODE")
	if node == "" {
		t.Skip("RENBROWSER_MEDIA_TEST_NODE not set; skipping live media test")
	}
	return node
}

// fetchPageRetry keeps retrying a page fetch until the node is reachable.
// Each attempt sends a path request, which both discovers the route and
// pulls the node's announce into the local announce handler when the
// shared daemon already knows the destination.
func fetchPageRetry(
	t *testing.T,
	stack *rns.Stack,
	nodeHash string,
	path string,
	timeout time.Duration,
) nomadnet.FetchResult {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last nomadnet.FetchResult
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		last = stack.Browser().Fetch(ctx, nodeHash, path, nomadnet.RequestData{})
		cancel()
		if last.Error == "" && len(last.Body) > 0 {
			return last
		}
		_, hasID := stack.Handler().Identity(nodeHash)
		t.Logf("fetch %s attempt failed: %s (announces=%d, nodeIdentity=%v)",
			path, last.Error, len(stack.Handler().List()), hasID)
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("could not fetch %s from %s within %s (last error: %s)", path, nodeHash, timeout, last.Error)
	return last
}

// startSharedInstanceStack attaches to the operator's real ~/.reticulum-go
// config, which is expected to point at a running shared rnsd instance.
// Unlike startLiveStack it does not isolate the data root or seed community
// uplinks: the media node under test is local to the shared daemon.
func startSharedInstanceStack(t *testing.T) *rns.Stack {
	t.Helper()
	if testing.Short() {
		t.Skip("skipped with -short")
	}
	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	if err := stack.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stack.Stop() })
	if mode := stack.SharedInstanceMode(); mode != "client" {
		t.Fatalf("expected shared instance client mode, got %q", mode)
	}
	return stack
}

func isWebP(body []byte) bool {
	return len(body) >= 12 &&
		string(body[0:4]) == "RIFF" &&
		string(body[8:12]) == "WEBP"
}

// TestLiveMediaWireFormat exercises the /media request convention against a
// real NomadNet node: the wire request path is "/media" and the payload is a
// {path, key} dict. A full-path request would never reach the handler.
func TestLiveMediaWireFormat(t *testing.T) {
	nodeHash := mediaTestNode(t)
	stack := startSharedInstanceStack(t)

	page := fetchPageRetry(t, stack, nodeHash, "/page/index.mu", 120*time.Second)
	t.Logf("page fetch ok: %d bytes", len(page.Body))

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	native := stack.Browser().FetchMedia(ctx, nodeHash, "/media/demo.webp", "", "", 0, nil)
	if native.Error != "" {
		t.Fatalf("native webp media fetch failed: %s", native.Error)
	}
	if !isWebP(native.Body) {
		t.Fatalf("native media response is not webp (%d bytes, magic %q)", len(native.Body), native.Body[:min(12, len(native.Body))])
	}
	t.Logf("native webp fetch ok: %d bytes, name=%q", len(native.Body), native.FileName)

	converted := stack.Browser().FetchMedia(ctx, nodeHash, "/media/test.png", "", "", 0, nil)
	if converted.Error != "" {
		t.Fatalf("converted media fetch failed: %s", converted.Error)
	}
	if !isWebP(converted.Body) {
		t.Fatalf("converted media response is not webp (%d bytes)", len(converted.Body))
	}
	t.Logf("png->webp converted fetch ok: %d bytes, name=%q", len(converted.Body), converted.FileName)
}

// TestLiveFetchNodeImage runs the full BrowserService path against the live
// node: media request, raster sniffing, disk image cache, and reload bypass.
func TestLiveFetchNodeImage(t *testing.T) {
	nodeHash := mediaTestNode(t)
	stack := startSharedInstanceStack(t)

	fetchPageRetry(t, stack, nodeHash, "/page/index.mu", 120*time.Second)

	root := t.TempDir()
	svc, err := app.NewBrowserServiceWithOptions(stack, nil, app.ServiceOptions{
		ProfilePath: filepath.Join(root, "profile.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(svc.Shutdown)

	rawURL := nodeHash + ":/media/demo.webp"
	first, err := svc.FetchNodeImage(rawURL, "", "", false)
	if err != nil {
		t.Fatalf("FetchNodeImage failed: %v", err)
	}
	if first.Mime != "image/webp" {
		t.Fatalf("expected image/webp, got %q", first.Mime)
	}
	if first.Bytes == 0 || first.Data == "" {
		t.Fatal("empty image result")
	}
	t.Logf("FetchNodeImage ok: %d bytes mime=%s name=%q", first.Bytes, first.Mime, first.Name)

	second, err := svc.FetchNodeImage(rawURL, "", "", false)
	if err != nil {
		t.Fatalf("cached FetchNodeImage failed: %v", err)
	}
	if second.Data != first.Data {
		t.Fatal("cached image bytes differ from first fetch")
	}

	entries, err := os.ReadDir(filepath.Join(root, "image-cache"))
	if err != nil {
		t.Fatalf("image cache dir missing: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("image cache dir is empty after fetch")
	}
	t.Logf("image cache holds %d entries", len(entries))

	reloaded, err := svc.FetchNodeImage(rawURL, "", "", true)
	if err != nil {
		t.Fatalf("reload FetchNodeImage failed: %v", err)
	}
	if reloaded.Bytes == 0 || reloaded.Data == "" {
		t.Fatal("empty image result on reload")
	}
}

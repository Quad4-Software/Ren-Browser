//go:build interop

// SPDX-License-Identifier: MIT

package interop_test

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"renbrowser/internal/limits"
	"renbrowser/internal/nomadnet"
	"renbrowser/internal/paths"
	"renbrowser/internal/rns"
)

func startLiveStack(t *testing.T) *rns.Stack {
	t.Helper()
	if testing.Short() {
		t.Skip("skipped with -short")
	}
	root := t.TempDir()
	paths.SetDataRoot(root)
	t.Cleanup(func() { paths.SetDataRoot("") })

	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	ensureInteropCommunityInterfaces(t, stack)
	if err := stack.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stack.Stop() })

	if len(waitForOutboundOnlineInterfaces(stack, 90*time.Second, 1)) == 0 {
		t.Fatal("no online outbound interfaces")
	}
	return stack
}

// TestLivePageBodiesRespectMaxPageBytes fetches real NomadNet pages over the
// public mesh and asserts Ren Browser's page cap holds for what arrived.
// RNS cannot size-filter encrypted payloads by design. The app must enforce.
func TestLivePageBodiesRespectMaxPageBytes(t *testing.T) {
	stack := startLiveStack(t)
	nodes := waitForNodes(stack.Handler(), 120*time.Second, 1)
	if len(nodes) == 0 {
		t.Fatal("no nodes discovered")
	}

	maxPage := limits.MaxPageBytes()
	browser := stack.Browser()
	got, ok := fetchFirstReachable(t, browser, nodes, 8, 2)
	if !ok {
		t.Fatal("could not fetch any discovered node")
	}
	if len(got.result.Body) > maxPage {
		t.Fatalf(
			"page body %d exceeds MaxPageBytes %d from %s",
			len(got.result.Body), maxPage, got.node.Hash,
		)
	}
	t.Logf(
		"page ok under cap: %s bytes=%d limit=%d hops=%d",
		got.node.Name, len(got.result.Body), maxPage, got.result.Hops,
	)

	// Probe a few more nodes. Any oversize body is a hard failure.
	checked := 1
	for _, node := range nodes {
		if node.Hash == got.node.Hash {
			continue
		}
		if checked >= 4 {
			break
		}
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		fetch := browser.Fetch(ctx, node.Hash, "/page/index.mu", nomadnet.RequestData{})
		cancel()
		checked++
		if fetch.Error != "" {
			t.Logf("skip %s: %s", node.Name, fetch.Error)
			continue
		}
		if len(fetch.Body) > maxPage {
			t.Fatalf("page body %d exceeds limit %d from %s", len(fetch.Body), maxPage, node.Hash)
		}
		t.Logf("page ok: %s bytes=%d", node.Name, len(fetch.Body))
	}
}

// TestLiveHeapDuringPageFetch samples heap around a live page fetch. This is
// a smoke check that a normal mesh page does not explode process memory.
func TestLiveHeapDuringPageFetch(t *testing.T) {
	stack := startLiveStack(t)
	nodes := waitForNodes(stack.Handler(), 120*time.Second, 1)
	if len(nodes) == 0 {
		t.Fatal("no nodes discovered")
	}

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	got, ok := fetchFirstReachable(t, stack.Browser(), nodes, 6, 2)
	if !ok {
		t.Fatal("could not fetch any discovered node")
	}

	runtime.ReadMemStats(&after)
	delta := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	t.Logf(
		"heap before=%d after=%d delta=%d body=%d node=%s",
		before.HeapAlloc, after.HeapAlloc, delta, len(got.result.Body), got.node.Name,
	)
	// Soft bound: a single index page plus stack overhead should stay well
	// under the page cap times a small multiplier even with render buffers.
	soft := int64(limits.MaxPageBytes()) * 8
	if delta > soft {
		t.Fatalf("heap delta %d exceeds soft bound %d after one page fetch", delta, soft)
	}
}

// TestLiveAnnounceMaxSizeKBObserved records peer-advertised page size hints.
// MaxSizeKB is advisory NomadNet metadata. Peers can lie. File downloads ignore
// it and remain unlimited by default.
func TestLiveAnnounceMaxSizeKBObserved(t *testing.T) {
	stack := startLiveStack(t)
	nodes := waitForNodes(stack.Handler(), 120*time.Second, 3)
	if len(nodes) == 0 {
		t.Fatal("no nodes discovered")
	}

	var withHint, zeroHint int
	for _, node := range nodes {
		if node.MaxSizeKB > 0 {
			withHint++
			t.Logf("announce MaxSizeKB=%d name=%s hash=%s hops=%d", node.MaxSizeKB, node.Name, node.Hash, node.Hops)
		} else {
			zeroHint++
		}
	}
	t.Logf("nodes=%d with MaxSizeKB hint=%d without=%d fileCap=%d (0=unlimited)", len(nodes), withHint, zeroHint, limits.MaxFileBytes())
	if limits.MaxFileBytes() != 0 {
		t.Fatal("interop expected default unlimited file policy unless env overridden")
	}
}

// TestLiveFileFetchHonoursConfiguredCap sets a small REN_BROWSER_MAX_FILE_BYTES
// and attempts /file/ fetches against reachable nodes. Success under the cap or
// an explicit too-large / empty / missing-file error is acceptable. A completed
// body larger than the cap is a hard failure.
func TestLiveFileFetchHonoursConfiguredCap(t *testing.T) {
	const fileCap = 64 * 1024
	t.Setenv("REN_BROWSER_MAX_FILE_BYTES", "65536")
	if limits.MaxFileBytes() != fileCap {
		t.Fatalf("MaxFileBytes=%d want %d", limits.MaxFileBytes(), fileCap)
	}

	stack := startLiveStack(t)
	nodes := waitForNodes(stack.Handler(), 180*time.Second, 1)
	if len(nodes) == 0 {
		t.Fatal("no nodes discovered")
	}
	// Keep collecting announces. High-hop ghosts are common on the public mesh.
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) && len(nodes) < 8 {
		time.Sleep(3 * time.Second)
		nodes = stack.Handler().List()
	}
	t.Logf("file-cap probe against %d discovered node(s)", len(nodes))

	got, ok := fetchFirstReachable(t, stack.Browser(), nodes, 12, 2)
	if !ok {
		t.Skip("no reachable NomadNet node for file-cap probe (mesh flakiness)")
	}

	// Common NomadNet file paths. Missing files are fine. Oversized must error.
	candidates := []string{
		"/file/index.mu",
		"/file/README",
		"/file/readme.txt",
		"/file/files",
	}
	// Pull simple /file/ names referenced from the page body when present.
	body := string(got.result.Body)
	for _, token := range []string{"/file/", "`/file/"} {
		idx := 0
		for {
			i := strings.Index(body[idx:], token)
			if i < 0 {
				break
			}
			i += idx
			start := i
			if strings.HasPrefix(body[i:], "`") {
				start++
			}
			end := start
			for end < len(body) {
				c := body[end]
				if c == '`' || c == ' ' || c == '\n' || c == '\r' || c == '|' || c == ')' {
					break
				}
				end++
			}
			path := body[start:end]
			if strings.HasPrefix(path, "/file/") && len(path) < 128 {
				candidates = append(candidates, path)
			}
			idx = end
			if idx >= len(body) {
				break
			}
		}
	}

	seen := map[string]bool{}
	browser := stack.Browser()
	attempts := 0
	for _, path := range candidates {
		if seen[path] {
			continue
		}
		seen[path] = true
		if attempts >= 6 {
			break
		}
		attempts++
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		fetch := browser.Fetch(ctx, got.node.Hash, path, nomadnet.RequestData{})
		cancel()
		if fetch.Error != "" {
			t.Logf("file %s: err=%s (ok if missing/oversized)", path, fetch.Error)
			if strings.Contains(strings.ToLower(fetch.Error), "too large") {
				t.Logf("cap enforced on %s", path)
			}
			continue
		}
		if len(fetch.Body) > fileCap {
			t.Fatalf("file %s returned %d bytes over cap %d", path, len(fetch.Body), fileCap)
		}
		t.Logf("file %s ok under cap: %d bytes", path, len(fetch.Body))
	}
	if attempts == 0 {
		t.Fatal("no file paths attempted")
	}
}

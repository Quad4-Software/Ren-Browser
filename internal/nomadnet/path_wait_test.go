// SPDX-License-Identifier: MIT
package nomadnet

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/identity"
	"quad4/reticulum-go/pkg/interfaces"
	"quad4/reticulum-go/pkg/transport"
)

func testPathTransport(t *testing.T) *transport.Transport {
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

func TestWaitPathRespectsContextCancel(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0x55}, 16)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := waitPath(ctx, tr, dest, 5*time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancel, got %v", err)
	}
}

func TestWaitPathInvalidDestination(t *testing.T) {
	tr := testPathTransport(t)
	err := waitPath(context.Background(), tr, []byte{1, 2, 3}, time.Second)
	if !errors.Is(err, errInvalidPathDestination) {
		t.Fatalf("expected invalid destination, got %v", err)
	}
}

func TestWaitPathTimesOut(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0x66}, 16)

	start := time.Now()
	err := waitPath(context.Background(), tr, dest, 250*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected timeout, got %v", err)
	}
	if elapsed := time.Since(start); elapsed < 200*time.Millisecond || elapsed > time.Second {
		t.Fatalf("unexpected timeout duration: %v", elapsed)
	}
}

func TestWaitPathReusesValidCachedPath(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0x77}, 16)
	udp, err := interfaces.NewUDPInterface("test-out", "127.0.0.1:0", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tr.RegisterInterface("test-out", udp); err != nil {
		t.Fatal(err)
	}
	tr.UpdatePath(dest, bytes.Repeat([]byte{0x22}, 16), "test-out", 1)
	if !tr.HasPath(dest) {
		t.Fatal("expected path")
	}
	if err := waitPath(context.Background(), tr, dest, time.Second); err != nil {
		t.Fatalf("expected reuse, got %v", err)
	}
}

func TestPathWaitError(t *testing.T) {
	if got := pathWaitError(context.DeadlineExceeded); got != "path discovery timed out" {
		t.Fatalf("got %q", got)
	}
}

func TestShouldRefreshPath(t *testing.T) {
	cases := map[string]bool{
		"no path to node: path discovery timed out": true,
		"link establish timeout":                    true,
		"node not discovered yet":                   true,
		"page not found":                            false,
		"":                                          false,
	}
	for msg, want := range cases {
		if got := ShouldRefreshPath(msg); got != want {
			t.Fatalf("ShouldRefreshPath(%q)=%v want %v", msg, got, want)
		}
	}
}

func TestRefreshPathForRetryExpires(t *testing.T) {
	tr := testPathTransport(t)
	b := NewBrowser(tr, NewAnnounceHandler())
	dest := bytes.Repeat([]byte{0x88}, 16)
	udp, err := interfaces.NewUDPInterface("ghost", "127.0.0.1:0", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tr.RegisterInterface("ghost", udp); err != nil {
		t.Fatal(err)
	}
	tr.UpdatePath(dest, bytes.Repeat([]byte{0x33}, 16), "ghost", 1)
	if !tr.HasPath(dest) {
		t.Fatal("expected path before refresh")
	}
	b.RefreshPathForRetry("88888888888888888888888888888888")
	if tr.HasPath(dest) {
		t.Fatal("expected path expired")
	}
}

func registerTestPath(t *testing.T, tr *transport.Transport, dest []byte, name string) {
	t.Helper()
	udp, err := interfaces.NewUDPInterface(name, "127.0.0.1:0", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tr.RegisterInterface(name, udp); err != nil {
		t.Fatal(err)
	}
	tr.UpdatePath(dest, bytes.Repeat([]byte{0x44}, 16), name, 1)
	if !tr.HasPath(dest) {
		t.Fatal("expected path")
	}
}

func TestWaitPathExpiresSoftStaleCachedPath(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0x99}, 16)
	registerTestPath(t, tr, dest, "soft-stale")

	prev := pathSoftMaxAge
	// Path table ages are Unix-second resolution; 0 means always expire.
	pathSoftMaxAge = 0
	t.Cleanup(func() { pathSoftMaxAge = prev })

	// Soft-stale entry must be dropped; without a live mesh this times out
	// instead of PrepareFreshReusedValidPath.
	err := waitPath(context.Background(), tr, dest, 120*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected rediscovery timeout after soft expire, got %v", err)
	}
	if tr.HasPath(dest) {
		t.Fatal("timed-out wait must leave path expired")
	}
}

func TestWaitPathKeepsFreshCachedPath(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0xaa}, 16)
	registerTestPath(t, tr, dest, "fresh-path")

	prev := pathSoftMaxAge
	pathSoftMaxAge = time.Hour
	t.Cleanup(func() { pathSoftMaxAge = prev })

	if err := waitPath(context.Background(), tr, dest, time.Second); err != nil {
		t.Fatalf("fresh path should reuse, got %v", err)
	}
	if !tr.HasPath(dest) {
		t.Fatal("fresh path must remain cached")
	}
}

func TestPrepareForWakeExpiresSoftStalePaths(t *testing.T) {
	tr := testPathTransport(t)
	b := NewBrowser(tr, NewAnnounceHandler())
	dest := bytes.Repeat([]byte{0xbb}, 16)
	registerTestPath(t, tr, dest, "wake-stale")
	// PrepareForWake only expires destinations it knows about (links/announces);
	// GetPathTable hashes are truncated and cannot be passed to ExpirePath.
	b.links[hex.EncodeToString(dest)] = nil

	prev := pathSoftMaxAge
	pathSoftMaxAge = 0
	t.Cleanup(func() { pathSoftMaxAge = prev })

	res := b.PrepareForWake()
	if res.ExpiredPaths < 1 {
		t.Fatalf("expected expired paths, got %+v", res)
	}
	if tr.HasPath(dest) {
		t.Fatal("soft-stale path should be gone after PrepareForWake")
	}
}

func TestPrepareForWakeKeepsFreshPaths(t *testing.T) {
	tr := testPathTransport(t)
	b := NewBrowser(tr, NewAnnounceHandler())
	dest := bytes.Repeat([]byte{0xcc}, 16)
	registerTestPath(t, tr, dest, "wake-fresh")

	prev := pathSoftMaxAge
	pathSoftMaxAge = time.Hour
	t.Cleanup(func() { pathSoftMaxAge = prev })

	res := b.PrepareForWake()
	if res.ExpiredPaths != 0 {
		t.Fatalf("fresh paths should not expire, got %+v", res)
	}
	if !tr.HasPath(dest) {
		t.Fatal("fresh path must remain")
	}
}

func TestPrepareForWakeDropsNilCachedLinks(t *testing.T) {
	tr := testPathTransport(t)
	b := NewBrowser(tr, NewAnnounceHandler())
	b.links["dead"] = nil
	res := b.PrepareForWake()
	if res.DroppedLinks < 1 {
		t.Fatalf("expected nil link drop, got %+v", res)
	}
	if _, ok := b.links["dead"]; ok {
		t.Fatal("nil link entry should be removed")
	}
}

func TestPathAgeMissing(t *testing.T) {
	tr := testPathTransport(t)
	if age := pathAge(tr, bytes.Repeat([]byte{0xdd}, 16)); age != -1 {
		t.Fatalf("missing path age=%v want -1", age)
	}
}

func TestPathAgeMatchesTruncatedTableHash(t *testing.T) {
	tr := testPathTransport(t)
	dest := bytes.Repeat([]byte{0xde}, 16)
	registerTestPath(t, tr, dest, "age-trunc")
	age := pathAge(tr, dest)
	if age < 0 {
		t.Fatal("pathAge must match GetPathTable truncated hashes")
	}
}

func TestPrepareForWakeExpiresAnnounceKnownSoftStalePath(t *testing.T) {
	tr := testPathTransport(t)
	handler := NewAnnounceHandler()
	b := NewBrowser(tr, handler)
	dest := bytes.Repeat([]byte{0xee}, 16)
	registerTestPath(t, tr, dest, "announce-wake")
	hash := hex.EncodeToString(dest)
	handler.mu.Lock()
	handler.nodes[hash] = announcedPeer{node: Node{Hash: hash, Name: "wake-node", Enabled: true}}
	handler.mu.Unlock()

	prev := pathSoftMaxAge
	pathSoftMaxAge = 0
	t.Cleanup(func() { pathSoftMaxAge = prev })

	res := b.PrepareForWake()
	if res.ExpiredPaths < 1 {
		t.Fatalf("expected announce-known path expired, got %+v", res)
	}
	if tr.HasPath(dest) {
		t.Fatal("soft-stale announce-known path should be gone")
	}
}

func TestCachedLinkRejectsIdleReuse(t *testing.T) {
	tr := testPathTransport(t)
	b := NewBrowser(tr, NewAnnounceHandler())
	dest := bytes.Repeat([]byte{0xef}, 16)
	key := hex.EncodeToString(dest)
	b.links[key] = nil
	if got := b.cachedLink(dest); got != nil {
		t.Fatal("nil cached link must not be reused")
	}
	if _, ok := b.links[key]; ok {
		t.Fatal("nil cached link entry should be removed by cachedLink")
	}
}

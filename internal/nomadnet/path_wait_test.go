// SPDX-License-Identifier: MIT
package nomadnet

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/identity"
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

func TestPathWaitError(t *testing.T) {
	if got := pathWaitError(context.DeadlineExceeded); got != "path discovery timed out" {
		t.Fatalf("got %q", got)
	}
}

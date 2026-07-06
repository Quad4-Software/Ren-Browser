// SPDX-License-Identifier: MIT
package nomadnet_test

import (
	"testing"

	"renbrowser/internal/nomadnet"
)

func TestCheckResponseSizeWithinLimit(t *testing.T) {
	if err := nomadnet.CheckResponseSize([]byte("ok"), 2, 8); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckResponseSizeOverReceived(t *testing.T) {
	err := nomadnet.CheckResponseSize(make([]byte, 10), 10, 8)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckResponseSizeOverAdvertised(t *testing.T) {
	err := nomadnet.CheckResponseSize([]byte("x"), 16, 8)
	if err == nil {
		t.Fatal("expected error")
	}
}

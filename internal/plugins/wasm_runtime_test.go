// SPDX-License-Identifier: MIT
package plugins

import (
	"testing"
)

func TestWasmRuntimeCloseIdempotent(t *testing.T) {
	rt := NewWasmRuntime()
	if err := rt.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := rt.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

func TestWasmRuntimeUnloadMissingModule(t *testing.T) {
	rt := NewWasmRuntime()
	rt.Unload("missing")
}

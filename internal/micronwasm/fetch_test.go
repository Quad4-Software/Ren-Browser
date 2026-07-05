// SPDX-License-Identifier: MIT
package micronwasm

import "testing"

func TestParseShasums256ForFilename(t *testing.T) {
	text := "abc123def4567890123456789012345678901234567890123456789012345678  micron-parser-go.wasm\n"
	want := "abc123def4567890123456789012345678901234567890123456789012345678"
	got, err := ParseShasums256ForFilename(text, WasmFilename)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestParseShasums256ForFilenameMissing(t *testing.T) {
	_, err := ParseShasums256ForFilename("deadbeef\n", WasmFilename)
	if err == nil {
		t.Fatal("expected error")
	}
}

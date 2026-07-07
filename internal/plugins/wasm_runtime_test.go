// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
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

func TestWasmRuntimeReloadPluginAfterUnload(t *testing.T) {
	wasmPath := translatorWasmPath(t)
	data, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}
	manifest := Manifest{
		ID:          "renbrowser.micron-translator",
		Name:        "Micron Translator",
		Version:     "1.0.0",
		Backend:     filepath.Base(wasmPath),
		Permissions: []string{PermNetworkFetch},
	}

	rt := NewWasmRuntime()
	t.Cleanup(func() { _ = rt.Close() })

	if _, err := rt.LoadPluginForTest(manifest.ID, data, manifest, nil); err != nil {
		t.Fatalf("first load: %v", err)
	}
	rt.Unload(manifest.ID)
	if _, err := rt.LoadPluginForTest(manifest.ID, data, manifest, nil); err != nil {
		t.Fatalf("reload after unload: %v", err)
	}
}

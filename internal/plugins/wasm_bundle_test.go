// SPDX-License-Identifier: MIT
package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/db"
	"renbrowser/internal/plugins"
)

func TestBundleWasmRoundTrip(t *testing.T) {
	t.Parallel()
	wasm := minimalWasmModule(t)
	manifest := plugins.Manifest{
		ManifestVersion: 1,
		ID:              "renbrowser.test-wasm",
		Name:            "Test WASM",
		Version:         "1.0.0",
		Main:            "main.js",
		Backend:         "plugin.wasm",
	}
	files := map[string]string{
		"main.js": "export function activate() {}",
	}
	bundled, err := plugins.BundleWasm(wasm, manifest, files)
	if err != nil {
		t.Fatalf("BundleWasm: %v", err)
	}
	parsed, err := plugins.ParseWasmBundle(bundled)
	if err != nil {
		t.Fatalf("ParseWasmBundle: %v", err)
	}
	if parsed.Manifest.ID != manifest.ID {
		t.Fatalf("manifest id = %q want %q", parsed.Manifest.ID, manifest.ID)
	}
	if parsed.Files["main.js"] != files["main.js"] {
		t.Fatalf("embedded main.js mismatch")
	}
	if err := parsed.ValidateEmbedded(); err != nil {
		t.Fatalf("ValidateEmbedded: %v", err)
	}
}

func TestInstallFromWasmBundledMicronTranslator(t *testing.T) {
	src := filepath.Join("..", "..", "extensions", "micron-translator")
	manifestPath := filepath.Join(src, plugins.ManifestFileName)
	wasmPath := filepath.Join(src, "translator.wasm")
	if _, err := os.Stat(wasmPath); err != nil {
		t.Skip("translator.wasm missing; run extensions/micron-translator/build-wasm.sh")
	}
	manifestRaw, err := os.ReadFile(manifestPath) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	var manifest plugins.Manifest
	if err := json.Unmarshal(manifestRaw, &manifest); err != nil {
		t.Fatal(err)
	}
	wasmData, err := os.ReadFile(wasmPath) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, name := range []string{"main.js", "settings.js"} {
		raw, err := os.ReadFile(filepath.Join(src, name)) // #nosec G304 -- test fixture path
		if err != nil {
			t.Fatal(err)
		}
		files[name] = string(raw)
	}
	bundled, err := plugins.BundleWasm(wasmData, manifest, files)
	if err != nil {
		t.Fatalf("BundleWasm: %v", err)
	}

	root := t.TempDir()
	bundlePath := filepath.Join(root, "micron-translator.wasm")
	if err := os.WriteFile(bundlePath, bundled, 0o640); err != nil {
		t.Fatal(err)
	}
	database, err := db.Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	manager := plugins.NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(root, "plugins"))
	installed, err := manager.InstallFromWasm(bundlePath, nil)
	if err != nil {
		t.Fatalf("InstallFromWasm: %v", err)
	}
	if installed.Manifest.ID != manifest.ID {
		t.Fatalf("installed id = %q", installed.Manifest.ID)
	}
	if !installed.Enabled {
		t.Fatal("expected plugin enabled after install")
	}
}

func TestParseWasmBundleRejectsInvalidMagic(t *testing.T) {
	t.Parallel()
	if _, err := plugins.ParseWasmBundle([]byte("not wasm")); err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func minimalWasmModule(t *testing.T) []byte {
	t.Helper()
	wasm := []byte(wasmMagic + wasmVersion)
	section := []byte{0}
	payload := append(encodeTestByteVector([]byte("test")), encodeTestByteVector([]byte("x"))...)
	section = append(section, encodeTestULEB32(uint32(len(payload)))...)
	section = append(section, payload...)
	return append(wasm, section...)
}

const wasmMagic = "\x00asm"
const wasmVersion = "\x01\x00\x00\x00"

func encodeTestULEB32(v uint32) []byte {
	var out []byte
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			b |= 0x80
		}
		out = append(out, b)
		if v == 0 {
			break
		}
	}
	return out
}

func encodeTestByteVector(data []byte) []byte {
	out := encodeTestULEB32(uint32(len(data)))
	return append(out, data...)
}

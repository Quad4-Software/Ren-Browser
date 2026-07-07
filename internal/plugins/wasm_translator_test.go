// SPDX-License-Identifier: MIT
package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func translatorWasmPath(t *testing.T) string {
	t.Helper()
	candidates := []string{
		filepath.Join("..", "..", "extensions", "micron-translator", "translator.wasm"),
		filepath.Join("testdata", "micron-translator", "translator.wasm"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	t.Skip("translator.wasm not built; run extensions/micron-translator/build-wasm.sh")
	return ""
}

func TestWasmTranslatorTranslatesMicronText(t *testing.T) {
	wasmPath := translatorWasmPath(t)
	data, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}

	rt := NewWasmRuntime()
	t.Cleanup(func() { _ = rt.Close() })

	manifest := Manifest{
		ID:          "renbrowser.micron-translator",
		Permissions: []string{PermNetworkFetch},
	}

	wp, err := rt.LoadPluginForTest("renbrowser.micron-translator", data, manifest, mockGoogleTranslate)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	req, err := json.Marshal(WasmTranslateRequest{
		Body: "Hello `>Title` world",
		Settings: WasmTranslateSettings{
			Backend:    "google",
			TargetLang: "es",
			SourceLang: "auto",
		},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	out, err := wp.CallExport("translate_micron", req)
	if err != nil {
		t.Fatalf("translate_micron: %v", err)
	}

	var resp WasmTranslateResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != "" {
		t.Fatalf("wasm error: %s", resp.Error)
	}
	if !strings.Contains(resp.Body, "Hola") {
		t.Fatalf("body = %q, want translated Spanish text", resp.Body)
	}
	if !strings.Contains(resp.Body, "`>Title`") {
		t.Fatalf("body = %q, want micron markup preserved", resp.Body)
	}
}

func TestWasmHostDeniesHTTPWithoutPermission(t *testing.T) {
	wasmPath := translatorWasmPath(t)
	data, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}

	rt := NewWasmRuntime()
	t.Cleanup(func() { _ = rt.Close() })

	manifest := Manifest{
		ID:          "renbrowser.micron-translator",
		Permissions: nil,
	}

	wp, err := rt.LoadPluginForTest("renbrowser.micron-translator", data, manifest, nil)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	req, err := json.Marshal(WasmTranslateRequest{
		Body: "Hello",
		Settings: WasmTranslateSettings{
			Backend:    "google",
			TargetLang: "es",
		},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	out, err := wp.CallExport("translate_micron", req)
	if err != nil {
		t.Fatalf("translate_micron: %v", err)
	}
	var resp WasmTranslateResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error == "" {
		t.Fatalf("expected permission error, got body=%q", resp.Body)
	}
}

func mockGoogleTranslate(req WasmHTTPRequest) (WasmHTTPResponse, error) {
	if strings.Contains(req.URL, "translate.googleapis.com") {
		return WasmHTTPResponse{
			StatusCode: 200,
			Body:       `)]}'\n[[["Hola","Hello",null,null,3]]]`,
		}, nil
	}
	return WasmHTTPResponse{StatusCode: 404, Body: "not found"}, nil
}

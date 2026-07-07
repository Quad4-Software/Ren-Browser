// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMicronTranslatorNetworkEndpoints(t *testing.T) {
	dir := filepath.Join("..", "..", "extensions", "micron-translator")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("micron-translator extension not present")
	}
	manifest, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	endpoints := CollectNetworkEndpoints(manifest, dir, nil)
	seen := make(map[string]struct{}, len(endpoints))
	for _, endpoint := range endpoints {
		seen[endpoint] = struct{}{}
	}
	want := []string{
		"https://translate.googleapis.com/",
		"https://libretranslate.com/",
	}
	for _, endpoint := range want {
		if _, ok := seen[endpoint]; !ok {
			t.Fatalf("missing endpoint %q in %#v", endpoint, endpoints)
		}
	}
	foundGoogleAPI := false
	for _, endpoint := range endpoints {
		if endpoint == "https://translate.googleapis.com/translate_a/single" ||
			endpoint == "https://translate.googleapis.com/translate_a/single?client=gtx&sl=" {
			foundGoogleAPI = true
		}
		if contains(endpoint, "translate.googleapis.com") && contains(endpoint, "translate_a") {
			foundGoogleAPI = true
		}
	}
	if !foundGoogleAPI {
		t.Fatalf("expected google translate API URL in scanned endpoints, got %#v", endpoints)
	}
	preview, err := PreviewInstallFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.I18nLocales) == 0 {
		t.Fatalf("expected bundled i18n locales, got %#v", preview.I18nLocales)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

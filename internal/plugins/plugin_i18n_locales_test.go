// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectPluginI18nLocalesFromDir(t *testing.T) {
	dir := t.TempDir()
	localesDir := filepath.Join(dir, "locales")
	if err := os.MkdirAll(localesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"en.json", "de.json"} {
		if err := os.WriteFile(filepath.Join(localesDir, name), []byte(`{"panel":{"title":"x"}}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got := CollectPluginI18nLocales(dir, nil)
	if len(got) != 2 || got[0] != "de" || got[1] != "en" {
		t.Fatalf("locales = %#v", got)
	}
}

func TestCollectPluginI18nLocalesFromEmbedded(t *testing.T) {
	embedded := map[string][]byte{
		"locales/en.json": []byte(`{}`),
		"locales/es.json": []byte(`{}`),
	}
	got := CollectPluginI18nLocales("", embedded)
	if len(got) != 2 || got[0] != "en" || got[1] != "es" {
		t.Fatalf("locales = %#v", got)
	}
}

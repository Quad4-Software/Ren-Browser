// SPDX-License-Identifier: MIT
package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/db"
)

func TestWasmCallBlocksTranslateWithoutNetworkPermission(t *testing.T) {
	wasmPath := translatorWasmPath(t)
	data, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm: %v", err)
	}

	tmp := t.TempDir()
	database, err := openTestDB(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	dest := filepath.Join(tmp, "plugins", "renbrowser.micron-translator")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"manifestVersion": 1,
		"id": "renbrowser.micron-translator",
		"name": "Micron Translator",
		"version": "1.0.0",
		"backend": "translator.wasm",
		"permissions": ["network.fetch"],
		"contributes": {}
	}`
	if err := os.WriteFile(filepath.Join(dest, ManifestFileName), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "translator.wasm"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	manager := NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	settings := RuntimeSettings{GrantedPermissions: nil}
	_ = manager.store.UpsertPlugin("renbrowser.micron-translator", true, settings.JSON())
	manager.mu.Lock()
	manager.plugins["renbrowser.micron-translator"] = InstalledPlugin{
		Manifest: Manifest{
			ID:          "renbrowser.micron-translator",
			Backend:     "translator.wasm",
			Permissions: []string{PermNetworkFetch},
		},
		Dir:                dest,
		Enabled:            true,
		GrantedPermissions: nil,
	}
	manager.mu.Unlock()
	if err := manager.enableLocked("renbrowser.micron-translator"); err != nil {
		t.Fatalf("enable: %v", err)
	}

	req, err := json.Marshal(WasmTranslateRequest{
		Body: "Hello",
		Settings: WasmTranslateSettings{
			Backend:    "google",
			TargetLang: "es",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.WasmCall("renbrowser.micron-translator", "translate_micron", string(req))
	if err == nil {
		t.Fatal("expected network permission error")
	}
}

func TestPluginHTTPFetchBudget(t *testing.T) {
	beginWasmFetchBudget("test.plugin", 2)
	t.Cleanup(func() { endWasmFetchBudget("test.plugin") })

	for i := range 2 {
		if err := consumeWasmFetchBudget("test.plugin"); err != nil {
			t.Fatalf("fetch %d: %v", i, err)
		}
	}
	if err := consumeWasmFetchBudget("test.plugin"); err == nil {
		t.Fatal("expected budget error")
	}
}

type nilStore struct{}

func (nilStore) UpsertPlugin(string, bool, string) error { return nil }
func (nilStore) GetPlugin(string) (db.PluginRow, error)  { return db.PluginRow{}, os.ErrNotExist }
func (nilStore) ListPlugins() ([]db.PluginRow, error)    { return nil, nil }
func (nilStore) SetPluginEnabled(string, bool) error     { return nil }
func (nilStore) DeletePlugin(string) error               { return nil }
func (nilStore) GetPluginSetting(string, string) (string, error) {
	return "", os.ErrNotExist
}
func (nilStore) SetPluginSetting(string, string, string) error { return nil }
func (nilStore) GetSetting(string) (string, error)             { return "", os.ErrNotExist }
func (nilStore) SetSetting(string, string) error               { return nil }

func openTestDB(path string) (*db.DB, error) {
	return db.Open(path)
}

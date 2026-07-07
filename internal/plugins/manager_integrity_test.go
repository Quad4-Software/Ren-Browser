// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/db"
)

func TestLoadAllDisablesTamperedExtension(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src", "renbrowser.tamper-test")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"manifestVersion": 1,
		"id": "renbrowser.tamper-test",
		"name": "Tamper test",
		"version": "1.0.0"
	}`
	if err := os.WriteFile(filepath.Join(srcDir, "renbrowser.plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	database, err := db.Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	manager := NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	if _, err := manager.InstallFromDir(srcDir, nil); err != nil {
		t.Fatalf("install: %v", err)
	}

	pluginDir := filepath.Join(tmp, "plugins", "renbrowser.tamper-test")
	if err := os.WriteFile(filepath.Join(pluginDir, "evil.js"), []byte("evil"), 0o644); err != nil {
		t.Fatal(err)
	}

	manager2 := NewManager(database)
	manager2.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	if err := manager2.LoadAll(); err != nil {
		t.Fatalf("load: %v", err)
	}
	p, ok := manager2.Get("renbrowser.tamper-test")
	if !ok {
		t.Fatal("plugin missing")
	}
	if p.Enabled {
		t.Fatal("expected tampered plugin to be disabled")
	}
	if !p.Tampered {
		t.Fatal("expected tampered flag")
	}
}

func TestEnableResetsIntegrityAfterTamper(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src", "renbrowser.tamper-reset")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"manifestVersion": 1,
		"id": "renbrowser.tamper-reset",
		"name": "Tamper reset",
		"version": "1.0.0"
	}`
	if err := os.WriteFile(filepath.Join(srcDir, "renbrowser.plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	database, err := db.Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	manager := NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	if _, err := manager.InstallFromDir(srcDir, nil); err != nil {
		t.Fatalf("install: %v", err)
	}

	pluginDir := filepath.Join(tmp, "plugins", "renbrowser.tamper-reset")
	if err := os.WriteFile(filepath.Join(pluginDir, "accepted.js"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	manager2 := NewManager(database)
	manager2.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	if err := manager2.LoadAll(); err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := manager2.Enable("renbrowser.tamper-reset"); err != nil {
		t.Fatalf("enable: %v", err)
	}
	p, ok := manager2.Get("renbrowser.tamper-reset")
	if !ok {
		t.Fatal("plugin missing")
	}
	if p.Tampered || !p.Enabled {
		t.Fatalf("expected enabled clean plugin, tampered=%v enabled=%v", p.Tampered, p.Enabled)
	}
}

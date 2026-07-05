// SPDX-License-Identifier: MIT
package plugins_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/db"
	"renbrowser/internal/plugins"
)

func TestInstallFromZip(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	zipPath := filepath.Join(tmp, "hello.renplugin.zip")
	if err := packHelloZip(zipPath); err != nil {
		t.Fatalf("pack zip: %v", err)
	}

	pluginsDir := filepath.Join(tmp, "plugins")
	manager := plugins.NewManager(database)
	manager.SetPluginsDirForTest(pluginsDir)

	installed, err := manager.InstallFromZip(zipPath)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if installed.Manifest.ID != "renbrowser.hello" {
		t.Fatalf("id = %q", installed.Manifest.ID)
	}
}

func TestInstallFromZipRejectsPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	zipPath := filepath.Join(tmp, "evil.renplugin.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	w := zip.NewWriter(f)
	entry, err := w.Create("../renbrowser.plugin.json")
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := entry.Write([]byte(`{"id":"evil"}`)); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	manager := plugins.NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	if _, err := manager.InstallFromZip(zipPath); err == nil {
		t.Fatal("expected path traversal zip to be rejected")
	}
}

func packHelloZip(dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	w := zip.NewWriter(f)
	for _, name := range []string{"renbrowser.plugin.json", "main.js"} {
		src := filepath.Join("testdata", "hello-extension", name)
		data, err := os.ReadFile(src)
		if err != nil {
			_ = w.Close()
			_ = f.Close()
			return err
		}
		entry, err := w.Create(name)
		if err != nil {
			_ = w.Close()
			_ = f.Close()
			return err
		}
		if _, err := entry.Write(data); err != nil {
			_ = w.Close()
			_ = f.Close()
			return err
		}
	}
	if err := w.Close(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

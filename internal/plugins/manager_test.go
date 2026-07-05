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

// SPDX-License-Identifier: MIT
package plugins_test

import (
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/db"
	"renbrowser/internal/plugins"
)

func TestInstallMicronTranslatorWithWasmBackend(t *testing.T) {
	src := filepath.Join("..", "..", "extensions", "micron-translator")
	if _, err := os.Stat(filepath.Join(src, "translator.wasm")); err != nil {
		t.Skip("translator.wasm missing; run extensions/micron-translator/build-wasm.sh")
	}

	tmp := t.TempDir()
	database, err := db.Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	manager := plugins.NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))

	installed, err := manager.InstallFromDir(src, nil)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if installed.Manifest.Backend != "translator.wasm" {
		t.Fatalf("backend = %q", installed.Manifest.Backend)
	}
	if !installed.Enabled {
		t.Fatal("expected plugin enabled after install")
	}
}

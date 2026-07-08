// SPDX-License-Identifier: MIT
package bootstrap_test

import (
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/bootstrap"
	"renbrowser/internal/brand"
	"renbrowser/internal/config"
	"renbrowser/internal/paths"
)

func TestHandleResetIfNeeded(t *testing.T) {
	tempRoot := t.TempDir()
	oldDataRoot := paths.DataRoot()
	paths.SetDataRoot(tempRoot)
	defer paths.SetDataRoot(oldDataRoot)

	dataDir := filepath.Join(tempRoot, brand.DataDirName)
	retConfigDir := filepath.Join(tempRoot, ".reticulum-go")

	setupDirs := func() {
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dataDir, "dummy_db.db"), []byte("db data"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(retConfigDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(retConfigDir, "config"), []byte("config data"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	setupDirs()
	cfg := config.Runtime{Reset: false}
	bootstrap.HandleResetIfNeeded(cfg)

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Error("expected browser data directory to exist when Reset is false")
	}
	if _, err := os.Stat(retConfigDir); os.IsNotExist(err) {
		t.Error("expected Reticulum config directory to exist when Reset is false")
	}

	cfg.Reset = true
	bootstrap.HandleResetIfNeeded(cfg)

	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Error("expected browser data directory to be removed when Reset is true")
	}
	if _, err := os.Stat(retConfigDir); !os.IsNotExist(err) {
		t.Error("expected Reticulum config directory to be removed when Reset is true")
	}

	setupDirs()
	customRetFile := filepath.Join(tempRoot, "custom-reticulum-config.conf")
	if err := os.WriteFile(customRetFile, []byte("custom config data"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg = config.Runtime{
		Reset:           true,
		ReticulumConfig: customRetFile,
	}

	bootstrap.HandleResetIfNeeded(cfg)

	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Error("expected browser data directory to be removed when Reset is true")
	}
	if _, err := os.Stat(customRetFile); !os.IsNotExist(err) {
		t.Error("expected custom Reticulum config file to be removed when Reset is true")
	}
	if _, err := os.Stat(retConfigDir); os.IsNotExist(err) {
		t.Error("expected default Reticulum config directory to remain intact when custom config is specified")
	}
}

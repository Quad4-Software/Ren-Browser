// SPDX-License-Identifier: MIT
package plugins_test

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/plugins"
)

func TestValidateManifestOK(t *testing.T) {
	m := plugins.Manifest{
		ManifestVersion: 1,
		ID:              "io.example.demo",
		Name:            "Demo",
		Version:         "1.0.0",
		Engines:         map[string]string{"renbrowser": ">=0.1.0"},
	}
	if err := plugins.ValidateManifest(m); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestValidateManifestRejectsBadID(t *testing.T) {
	m := plugins.Manifest{ID: "bad id", Name: "Demo", Version: "1.0.0"}
	if err := plugins.ValidateManifest(m); err == nil {
		t.Fatal("expected invalid id error")
	}
}

func TestLoadHelloFixture(t *testing.T) {
	dir := filepath.Join("testdata", "hello-extension")
	m, err := plugins.LoadManifest(dir)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if m.ID != "renbrowser.hello" {
		t.Fatalf("id = %q", m.ID)
	}
}

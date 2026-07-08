// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCollectNetworkEndpointsManifestAndScan(t *testing.T) {
	dir := t.TempDir()
	manifest := Manifest{
		ID:   "test.plugin",
		Name: "Test",
		Network: &PluginNetwork{
			Endpoints: []string{
				"https://api.example.com/v1/",
				"User-configured LibreTranslate instance",
			},
		},
		Permissions: []string{PermNetworkFetch},
	}
	if err := os.WriteFile(filepath.Join(dir, "main.js"), []byte(`
const google = "https://translate.googleapis.com/translate_a/single";
const local = "http://localhost:8080/ignore";
`), 0o600); err != nil {
		t.Fatal(err)
	}

	endpoints := CollectNetworkEndpoints(manifest, dir, nil)
	if len(endpoints) < 3 {
		t.Fatalf("endpoints = %#v, want at least 3 entries", endpoints)
	}
	if endpoints[0] != "https://api.example.com/v1/" {
		t.Fatalf("first endpoint = %q", endpoints[0])
	}
	if endpoints[1] != "User-configured LibreTranslate instance" {
		t.Fatalf("manifest endpoint = %q", endpoints[1])
	}
	foundGoogle := slices.Contains(endpoints, "https://translate.googleapis.com/translate_a/single")
	if !foundGoogle {
		t.Fatalf("scanned google endpoint missing from %#v", endpoints)
	}
}

func TestPreviewInstallFromDirRequiresNetworkFetch(t *testing.T) {
	dir := t.TempDir()
	manifest := `{
  "manifestVersion": 1,
  "id": "test.network",
  "name": "Network",
  "version": "1.0.0",
  "permissions": ["network.fetch"],
  "network": { "endpoints": ["https://example.com/"] },
  "contributes": {}
}`
	if err := os.WriteFile(filepath.Join(dir, ManifestFileName), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}

	preview, err := PreviewInstallFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.RequiresNetworkFetch {
		t.Fatal("expected network.fetch requirement")
	}
	if len(preview.NetworkEndpoints) != 1 || preview.NetworkEndpoints[0] != "https://example.com/" {
		t.Fatalf("endpoints = %#v", preview.NetworkEndpoints)
	}
}

// SPDX-License-Identifier: MIT
package plugins

import "testing"

func TestNormalizeGrantedPermissionsDefaults(t *testing.T) {
	manifest := Manifest{
		ID:          "test.perms",
		Name:        "Test",
		Version:     "1.0.0",
		Permissions: []string{"storage.plugin", "network.fetch"},
	}
	got := NormalizeGrantedPermissions(manifest, nil)
	if len(got) != 2 {
		t.Fatalf("granted = %#v", got)
	}
}

func TestNormalizeGrantedPermissionsSubset(t *testing.T) {
	manifest := Manifest{
		ID:          "test.perms",
		Name:        "Test",
		Version:     "1.0.0",
		Permissions: []string{"storage.plugin", "network.fetch"},
	}
	got := NormalizeGrantedPermissions(manifest, []string{"storage.plugin", "unknown"})
	if len(got) != 1 || got[0] != "storage.plugin" {
		t.Fatalf("granted = %#v", got)
	}
}

func TestExtractURLsFromManifestText(t *testing.T) {
	endpoints := CollectNetworkEndpoints(Manifest{
		Network: &PluginNetwork{
			Endpoints: []string{
				"https://translate.googleapis.com/",
				"https://libretranslate.com/",
				"User-configured LibreTranslate instance URL",
			},
		},
	}, "", nil)
	if len(endpoints) != 3 {
		t.Fatalf("endpoints = %#v", endpoints)
	}
}

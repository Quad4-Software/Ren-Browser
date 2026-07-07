// SPDX-License-Identifier: MIT
package app_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/app"
	"renbrowser/internal/db"
	"renbrowser/internal/plugins"
)

func TestPluginFetchRequiresNetworkPermission(t *testing.T) {
	tmp := t.TempDir()
	database, err := db.Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	manager := plugins.NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	host := app.NewPluginHost(manager)

	_, err = host.PluginFetch("renbrowser.micron-translator", app.PluginHTTPRequest{
		Method: http.MethodGet,
		URL:    "https://example.com/",
	})
	if err == nil {
		t.Fatal("expected permission error")
	}
}

func TestPluginFetchGET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ok" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src", "renbrowser.fetch-test")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"manifestVersion": 1,
		"id": "renbrowser.fetch-test",
		"name": "Fetch test",
		"version": "1.0.0",
		"permissions": ["network.fetch"]
	}`
	if err := os.WriteFile(filepath.Join(srcDir, "renbrowser.plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	database, err := db.Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	manager := plugins.NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	if _, err := manager.InstallFromDir(srcDir, nil); err != nil {
		t.Fatalf("install: %v", err)
	}
	if err := manager.Enable("renbrowser.fetch-test"); err != nil {
		t.Fatalf("enable: %v", err)
	}

	host := app.NewPluginHost(manager)
	resp, err := host.PluginFetch("renbrowser.fetch-test", app.PluginHTTPRequest{
		Method: http.MethodGet,
		URL:    server.URL + "/ok",
	})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if resp.Body != `{"ok":true}` {
		t.Fatalf("body = %q", resp.Body)
	}
}

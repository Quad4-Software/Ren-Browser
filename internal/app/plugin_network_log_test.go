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

func TestPluginFetchRecordsNetworkLog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src", "renbrowser.fetch-log")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"manifestVersion": 1,
		"id": "renbrowser.fetch-log",
		"name": "Fetch log test",
		"version": "1.0.0",
		"permissions": ["network.fetch"]
	}`
	if err := os.WriteFile(filepath.Join(srcDir, "renbrowser.plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	database, err := db.Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	manager := plugins.NewManager(database)
	manager.SetPluginsDirForTest(filepath.Join(tmp, "plugins"))
	svc := &app.BrowserService{}
	manager.SetNetworkRecorder(svc.RecordPluginNetworkFetch)
	if _, err := manager.InstallFromDir(srcDir, nil); err != nil {
		t.Fatalf("install: %v", err)
	}

	host := app.NewPluginHost(manager)
	if _, err := host.PluginFetch("renbrowser.fetch-log", app.PluginHTTPRequest{
		Method: http.MethodGet,
		URL:    server.URL,
	}); err != nil {
		t.Fatalf("fetch: %v", err)
	}

	log := svc.GetNetworkLog()
	if len(log) != 1 {
		t.Fatalf("network log len = %d", len(log))
	}
	entry := log[0]
	if entry.Source != "plugin" || entry.PluginID != "renbrowser.fetch-log" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
	if entry.URL != server.URL {
		t.Fatalf("url = %q", entry.URL)
	}
}

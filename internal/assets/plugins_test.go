// SPDX-License-Identifier: MIT
package assets_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/assets"
	"renbrowser/internal/plugins"
)

type stubProvider struct {
	root string
}

func (s stubProvider) PluginRoot(id string) (string, bool) {
	if id != "demo" {
		return "", false
	}
	return s.root, true
}

func TestPluginHandlerServesFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.js"), []byte("export {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	handler := assets.PluginHandler(stubProvider{root: root}, http.NotFoundHandler())
	req := httptest.NewRequest(http.MethodGet, "/_plugins/demo/main.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if body := rec.Body.String(); body != "export {}" {
		t.Fatalf("body = %q", body)
	}
}

func TestPluginHandlerRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	handler := assets.PluginHandler(stubProvider{root: root}, http.NotFoundHandler())
	req := httptest.NewRequest(http.MethodGet, "/_plugins/demo/../secret", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestManagerImplementsProvider(t *testing.T) {
	var _ assets.PluginAssetProvider = (*plugins.Manager)(nil)
}

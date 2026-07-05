// SPDX-License-Identifier: MIT
package assets

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PluginAssetProvider interface {
	PluginRoot(id string) (string, bool)
}

func PluginHandler(provider PluginAssetProvider, next http.Handler) http.Handler {
	if provider == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/_plugins/") {
			next.ServeHTTP(w, r)
			return
		}
		trimmed := strings.TrimPrefix(r.URL.Path, "/_plugins/")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			http.NotFound(w, r)
			return
		}
		pluginID := parts[0]
		rel := filepath.Clean("/" + parts[1])
		if strings.HasPrefix(rel, "..") {
			http.NotFound(w, r)
			return
		}
		rel = strings.TrimPrefix(rel, "/")
		root, ok := provider.PluginRoot(pluginID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		full := filepath.Join(root, rel)
		if !strings.HasPrefix(filepath.Clean(full), filepath.Clean(root)) {
			http.NotFound(w, r)
			return
		}
		data, err := os.ReadFile(full) // #nosec G304 G703 -- jailed to plugin root
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctype := mime.TypeByExtension(filepath.Ext(full))
		if ctype == "" {
			switch filepath.Ext(full) {
			case ".js", ".mjs":
				ctype = "application/javascript"
			case ".wasm":
				ctype = "application/wasm"
			default:
				ctype = "application/octet-stream"
			}
		}
		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(data) // #nosec G705 -- plugin assets are user-installed extension content
	})
}

// SPDX-License-Identifier: MIT
package servermw

import (
	"net/http"
	"strings"

	"renbrowser/internal/brand"
)

type Options struct {
	TrustProxy bool
	BasePath   string
}

func Wrap(handler http.Handler, opts Options) http.Handler {
	h := handler
	if base := normalizeBasePath(opts.BasePath); base != "" {
		h = stripBasePath(base, h)
	}
	if opts.TrustProxy {
		h = trustForwardedHeaders(h)
	}
	return h
}

func normalizeBasePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	return "/" + path
}

func stripBasePath(base string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, base) {
			http.NotFound(w, r)
			return
		}
		trimmed := strings.TrimPrefix(r.URL.Path, base)
		if trimmed == "" {
			trimmed = "/"
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path = trimmed
		next.ServeHTTP(w, r2)
	})
}

func trustForwardedHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			r.URL.Scheme = proto
		}
		if host := r.Header.Get("X-Forwarded-Host"); host != "" {
			r.Host = host
		}
		if prefix := strings.TrimSpace(r.Header.Get("X-Forwarded-Prefix")); prefix != "" {
			r.Header.Set(brand.ProxyHeader, normalizeBasePath(prefix))
		}
		next.ServeHTTP(w, r)
	})
}

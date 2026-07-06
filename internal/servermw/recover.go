// SPDX-License-Identifier: MIT
package servermw

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
)

func sanitizeLogField(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func recoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error(
					"http handler panic",
					"method", sanitizeLogField(r.Method),
					"path", sanitizeLogField(r.URL.Path),
					"panic", rec,
					"stack", string(debug.Stack()),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func readyHandler(next http.Handler, ready func() bool) http.Handler {
	if ready == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" {
			if ready() {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
				return
			}
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}

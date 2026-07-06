//go:build server

// SPDX-License-Identifier: MIT
package serverauth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"renbrowser/internal/config"
)

func FuzzGuardLogin(f *testing.F) {
	f.Add(`{"password":"x"}`)
	f.Add(`{"password":""}`)
	f.Add(`{"password":"secret"}`)
	f.Add(`not json`)
	f.Add("")

	f.Fuzz(func(t *testing.T, body string) {
		if len(body) > 8192 {
			t.Skip("skip oversized body")
		}
		guard, _ := newTestGuard(t, config.Runtime{AuthBruteMax: 100})
		handler := guard.Middleware()(okHandler())

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "fuzz-agent")
		req.RemoteAddr = "198.51.100.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		switch rec.Code {
		case http.StatusOK, http.StatusBadRequest, http.StatusUnauthorized, http.StatusTooManyRequests:
		default:
			t.Fatalf("unexpected status %d for body len %d", rec.Code, len(body))
		}
	})
}

func FuzzGuardRoute(f *testing.F) {
	f.Add("/")
	f.Add("/api/auth/status")
	f.Add("/wails/runtime")
	f.Add("/../wails/runtime")
	f.Add("/ren/api/auth/login")

	f.Fuzz(func(t *testing.T, route string) {
		if len(route) > 512 || !strings.HasPrefix(route, "/") {
			t.Skip("skip invalid route")
		}
		guard, _ := newTestGuard(t, config.Runtime{BasePath: "/ren"})
		handler := guard.Middleware()(okHandler())

		req := httptest.NewRequest(http.MethodGet, route, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code < 200 || rec.Code >= 600 {
			t.Fatalf("invalid status %d for %q", rec.Code, route)
		}
	})
}

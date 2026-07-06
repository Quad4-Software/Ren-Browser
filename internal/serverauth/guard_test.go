//go:build server

// SPDX-License-Identifier: MIT
package serverauth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"renbrowser/internal/auth"
	"renbrowser/internal/config"
	"renbrowser/internal/db"
	"renbrowser/internal/serverauth"
)

func newTestGuard(t *testing.T, cfg config.Runtime) (*serverauth.Guard, *db.DB) {
	t.Helper()
	database, err := openTestDB(t)
	if err != nil {
		t.Fatal(err)
	}
	return newGuardWithCredential(t, database, cfg)
}

func openTestDB(t *testing.T) (*db.DB, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "guard.db")
	database, err := db.Open(path)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = database.Close() })
	return database, nil
}

func newGuardWithCredential(t *testing.T, database *db.DB, cfg config.Runtime) (*serverauth.Guard, *db.DB) {
	t.Helper()
	hash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := database.SetAuthCredential(hash); err != nil {
		t.Fatal(err)
	}
	guard, err := serverauth.NewGuard(database, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return guard, database
}

func TestGuardBlocksWailsWithoutSession(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	mw := guard.Middleware()
	next := httptest.NewRecorder()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/wails/runtime", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
	_ = next
}

func TestGuardLoginAndSession(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	mw := guard.Middleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body, _ := json.Marshal(map[string]string{"password": "secret"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "renbrowser_session" {
		t.Fatal("expected session cookie")
	}

	runtimeReq := httptest.NewRequest(http.MethodPost, "/wails/runtime", nil)
	runtimeReq.AddCookie(cookies[0])
	runtimeRec := httptest.NewRecorder()
	handler.ServeHTTP(runtimeRec, runtimeReq)
	if runtimeRec.Code != http.StatusOK {
		t.Fatalf("runtime status=%d", runtimeRec.Code)
	}
}

func TestGuardBruteForceBan(t *testing.T) {
	cfg := config.Runtime{AuthBruteMax: 3, AuthBruteBanMin: 15}
	guard, _ := newTestGuard(t, cfg)
	mw := guard.Middleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		body, _ := json.Marshal(map[string]string{"password": "wrong"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "brute-bot")
		req.RemoteAddr = "203.0.113.10:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(`{"password":"wrong"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "brute-bot")
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestGuardWhitelistBypass(t *testing.T) {
	cfg := config.Runtime{AuthIPWhitelist: "127.0.0.1"}
	guard, _ := newTestGuard(t, cfg)
	mw := guard.Middleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/wails/runtime", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

//go:build server

// SPDX-License-Identifier: MIT
package serverauth_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"renbrowser/internal/config"
	"renbrowser/internal/serverauth"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func loginRequest(t *testing.T, handler http.Handler, password string, opts ...func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "security-test")
	req.RemoteAddr = "198.51.100.20:1234"
	for _, opt := range opts {
		opt(req)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func sessionCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == "renbrowser_session" {
			return c
		}
	}
	return nil
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func TestGuardRejectsInvalidSessionCookie(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/wails/runtime", nil)
	req.AddCookie(&http.Cookie{Name: "renbrowser_session", Value: "not-a-real-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestGuardRejectsExpiredSession(t *testing.T) {
	guard, database := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	loginRec := loginRequest(t, handler, "secret")
	cookie := sessionCookie(loginRec)
	if cookie == nil {
		t.Fatal("missing session cookie")
	}

	tokenHash := hashSessionToken(cookie.Value)
	if err := database.DeleteAuthSession(tokenHash); err != nil {
		t.Fatal(err)
	}
	if err := database.CreateAuthSession(tokenHash, time.Now().Add(-time.Hour).Unix()); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/wails/runtime", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestGuardSessionCookieHttpOnly(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	rec := loginRequest(t, guard.Middleware()(okHandler()), "secret")
	cookie := sessionCookie(rec)
	if cookie == nil {
		t.Fatal("missing cookie")
	}
	if !cookie.HttpOnly {
		t.Fatal("session cookie must be HttpOnly")
	}
	if cookie.Value == "" {
		t.Fatal("empty session value")
	}
}

func TestGuardLoginIssuesUniqueSessions(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	a := sessionCookie(loginRequest(t, handler, "secret"))
	b := sessionCookie(loginRequest(t, handler, "secret"))
	if a == nil || b == nil {
		t.Fatal("missing cookies")
	}
	if a.Value == b.Value {
		t.Fatal("expected unique session tokens per login")
	}
}

func TestGuardLogoutRevokesSession(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	loginRec := loginRequest(t, handler, "secret")
	cookie := sessionCookie(loginRec)
	if cookie == nil {
		t.Fatal("missing cookie")
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(cookie)
	logoutRec := httptest.NewRecorder()
	handler.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusOK {
		t.Fatalf("logout status=%d", logoutRec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/wails/runtime", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestGuardRejectsOversizedLoginBody(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	body := strings.Repeat("x", 5000)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestGuardRejectsInvalidLoginJSON(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestGuardAuthEndpointMethods(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	cases := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/api/auth/login", http.StatusMethodNotAllowed},
		{http.MethodGet, "/api/auth/logout", http.StatusMethodNotAllowed},
		{http.MethodPost, "/api/auth/status", http.StatusMethodNotAllowed},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("%s %s status=%d", tc.method, tc.path, rec.Code)
		}
	}
}

func TestGuardPublicPathsWithoutSession(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	for _, path := range []string{"/", "/index.html", "/app.js", "/assets/logo.svg", "/health"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d", path, rec.Code)
		}
	}
}

func TestGuardBlocksProtectedAPIWithoutSession(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	for _, path := range []string{"/wails/runtime", "/wails/events", "/api/other"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s status=%d", path, rec.Code)
		}
	}
}

func TestGuardBasePath(t *testing.T) {
	cfg := config.Runtime{BasePath: "/ren"}
	guard, _ := newTestGuard(t, cfg)
	handler := guard.Middleware()(okHandler())

	body, _ := json.Marshal(map[string]string{"password": "secret"})
	req := httptest.NewRequest(http.MethodPost, "/ren/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGuardTrustedClientRelaxedBruteLimit(t *testing.T) {
	cfg := config.Runtime{AuthBruteMax: 2, AuthBruteRelax: 5, AuthBruteBanMin: 15}
	guard, _ := newTestGuard(t, cfg)
	handler := guard.Middleware()(okHandler())

	loginRequest(t, handler, "secret")

	for i := 0; i < 4; i++ {
		rec := loginRequest(t, handler, "wrong")
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("trusted client banned after %d failures", i+1)
		}
	}
}

func TestGuardNoBruteForceDisablesBan(t *testing.T) {
	cfg := config.Runtime{AuthBruteMax: 2, AuthBruteBanMin: 15, NoBruteForce: true}
	guard, _ := newTestGuard(t, cfg)
	handler := guard.Middleware()(okHandler())

	for i := 0; i < 6; i++ {
		rec := loginRequest(t, handler, "wrong", func(r *http.Request) {
			r.Header.Set("User-Agent", "no-ban-agent")
			r.RemoteAddr = "203.0.113.99:1234"
		})
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("unexpected ban on attempt %d", i+1)
		}
	}
}

func TestGuardBruteForceIsolatedByUserAgent(t *testing.T) {
	cfg := config.Runtime{AuthBruteMax: 2, AuthBruteBanMin: 15}
	guard, _ := newTestGuard(t, cfg)
	handler := guard.Middleware()(okHandler())

	banAgent := func(ua string) bool {
		for i := 0; i < 3; i++ {
			rec := loginRequest(t, handler, "wrong", func(r *http.Request) {
				r.Header.Set("User-Agent", ua)
				r.RemoteAddr = "203.0.113.50:1234"
			})
			if rec.Code == http.StatusTooManyRequests {
				return true
			}
		}
		return false
	}

	if !banAgent("blocked-agent") {
		t.Fatal("expected first agent to be banned")
	}
	rec := loginRequest(t, handler, "wrong", func(r *http.Request) {
		r.Header.Set("User-Agent", "other-agent")
		r.RemoteAddr = "203.0.113.50:1234"
	})
	if rec.Code == http.StatusTooManyRequests {
		t.Fatal("same IP different UA should not inherit ban")
	}
}

func TestGuardConcurrentFailedLogins(t *testing.T) {
	cfg := config.Runtime{AuthBruteMax: 100, AuthBruteBanMin: 15}
	guard, _ := newTestGuard(t, cfg)
	handler := guard.Middleware()(okHandler())

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			loginRequest(t, handler, "wrong", func(r *http.Request) {
				r.Header.Set("User-Agent", "concurrent")
				r.RemoteAddr = "198.51.100.77:1234"
			})
		}(i)
	}
	wg.Wait()
}

func TestGuardStatusReflectsSession(t *testing.T) {
	guard, _ := newTestGuard(t, config.Runtime{})
	handler := guard.Middleware()(okHandler())

	statusReq := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	statusRec := httptest.NewRecorder()
	handler.ServeHTTP(statusRec, statusReq)
	var before map[string]any
	if err := json.Unmarshal(statusRec.Body.Bytes(), &before); err != nil {
		t.Fatal(err)
	}
	if before["authenticated"] != false {
		t.Fatalf("before=%v", before)
	}

	loginRec := loginRequest(t, handler, "secret")
	cookie := sessionCookie(loginRec)

	statusReq = httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	statusReq.AddCookie(cookie)
	statusRec = httptest.NewRecorder()
	handler.ServeHTTP(statusRec, statusReq)
	var after map[string]any
	if err := json.Unmarshal(statusRec.Body.Bytes(), &after); err != nil {
		t.Fatal(err)
	}
	if after["authenticated"] != true {
		t.Fatalf("after=%v", after)
	}
}

func TestGuardDisabledPassesThrough(t *testing.T) {
	database, err := openTestDB(t)
	if err != nil {
		t.Fatal(err)
	}
	guard, err := serverauth.NewGuard(database, config.Runtime{})
	if err != nil {
		t.Fatal(err)
	}
	if guard.Enabled() {
		t.Fatal("expected disabled guard")
	}
	handler := guard.Middleware()(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/wails/runtime", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

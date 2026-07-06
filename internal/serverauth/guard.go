//go:build server

// SPDX-License-Identifier: MIT
package serverauth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"renbrowser/internal/auth"
	"renbrowser/internal/config"
	"renbrowser/internal/db"
)

const (
	sessionCookieName = "renbrowser_session"
	apiLoginPath      = "/api/auth/login"
	apiLogoutPath     = "/api/auth/logout"
	apiStatusPath     = "/api/auth/status"
)

type Guard struct {
	db              *db.DB
	enabled         bool
	trustProxy      bool
	basePath        string
	pepper          string
	bruteEnabled    bool
	bruteMax        int
	bruteRelaxedMax int
	bruteBan        time.Duration
	sessionTTL      time.Duration
	whitelist       *auth.Whitelist
}

type StatusResponse struct {
	AuthRequired  bool `json:"authRequired"`
	Authenticated bool `json:"authenticated"`
}

type loginRequest struct {
	Password string `json:"password"`
}

type loginResponse struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Blocked bool   `json:"blocked,omitempty"`
	RetryIn int64  `json:"retryIn,omitempty"`
}

func NewGuard(database *db.DB, cfg config.Runtime) (*Guard, error) {
	if database == nil {
		return nil, errors.New("database required")
	}
	enabled, err := database.AuthEnabled()
	if err != nil {
		return nil, err
	}
	if !enabled {
		return &Guard{db: database, enabled: false}, nil
	}

	pepper, err := loadOrCreatePepper(database)
	if err != nil {
		return nil, err
	}

	var whitelist *auth.Whitelist
	if strings.TrimSpace(cfg.AuthIPWhitelist) != "" {
		whitelist, err = auth.ParseWhitelist(cfg.AuthIPWhitelist)
		if err != nil {
			return nil, fmt.Errorf("auth ip whitelist: %w", err)
		}
	}

	bruteMax := cfg.AuthBruteMax
	if bruteMax <= 0 {
		bruteMax = 3
	}
	bruteRelax := cfg.AuthBruteRelax
	if bruteRelax <= 0 {
		bruteRelax = 10
	}
	banMin := cfg.AuthBruteBanMin
	if banMin <= 0 {
		banMin = 15
	}
	sessionHrs := cfg.AuthSessionHrs
	if sessionHrs <= 0 {
		sessionHrs = 168
	}

	return &Guard{
		db:              database,
		enabled:         true,
		trustProxy:      cfg.TrustProxy,
		basePath:        normalizeBasePath(cfg.BasePath),
		pepper:          pepper,
		bruteEnabled:    !cfg.NoBruteForce,
		bruteMax:        bruteMax,
		bruteRelaxedMax: bruteRelax,
		bruteBan:        time.Duration(banMin) * time.Minute,
		sessionTTL:      time.Duration(sessionHrs) * time.Hour,
		whitelist:       whitelist,
	}, nil
}

func (g *Guard) Enabled() bool {
	return g != nil && g.enabled
}

func (g *Guard) Activate() {
	if g != nil {
		g.enabled = true
	}
}

func (g *Guard) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if g == nil || !g.enabled {
				next.ServeHTTP(w, r)
				return
			}
			g.serve(w, r, next)
		})
	}
}

func (g *Guard) serve(w http.ResponseWriter, r *http.Request, next http.Handler) {
	_ = g.db.PruneExpiredAuthSessions()

	route := stripBasePath(r.URL.Path, g.basePath)
	ip := auth.ClientIP(r, g.trustProxy)
	if g.whitelist != nil && g.whitelist.Allows(ip) {
		next.ServeHTTP(w, r)
		return
	}

	switch route {
	case apiStatusPath:
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		g.writeStatus(w, r)
		return
	case apiLoginPath:
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		g.handleLogin(w, r)
		return
	case apiLogoutPath:
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		g.handleLogout(w, r)
		return
	}

	if g.hasValidSession(r) {
		next.ServeHTTP(w, r)
		return
	}

	if isPublicPath(route) || route == "/health" {
		next.ServeHTTP(w, r)
		return
	}

	if strings.HasPrefix(route, "/wails/") || strings.HasPrefix(route, "/api/") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	next.ServeHTTP(w, r)
}

func (g *Guard) writeStatus(w http.ResponseWriter, r *http.Request) {
	resp := StatusResponse{
		AuthRequired:  true,
		Authenticated: g.hasValidSession(r),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (g *Guard) handleLogin(w http.ResponseWriter, r *http.Request) {
	ip := auth.ClientIP(r, g.trustProxy)
	ua := auth.ClientUserAgent(r)
	clientHash := auth.ClientHash(g.pepper, ip, ua)

	if blocked, retryIn := g.isBlocked(clientHash); blocked {
		writeJSON(w, http.StatusTooManyRequests, loginResponse{
			OK:      false,
			Error:   "too many failed attempts",
			Blocked: true,
			RetryIn: retryIn,
		})
		return
	}

	var req loginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, loginResponse{OK: false, Error: "invalid request"})
		return
	}

	cred, err := g.db.GetAuthCredential()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, loginResponse{OK: false, Error: "auth unavailable"})
		return
	}
	if err := auth.VerifyPassword(cred.PasswordHash, req.Password); err != nil {
		g.recordFailure(clientHash)
		writeJSON(w, http.StatusUnauthorized, loginResponse{OK: false, Error: "invalid password"})
		return
	}

	g.recordSuccess(clientHash)
	token, err := auth.NewSessionToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, loginResponse{OK: false, Error: "session error"})
		return
	}
	expiresAt := time.Now().Add(g.sessionTTL).Unix()
	if err := g.db.CreateAuthSession(hashToken(token), expiresAt); err != nil {
		writeJSON(w, http.StatusInternalServerError, loginResponse{OK: false, Error: "session error"})
		return
	}
	g.setSessionCookie(w, token, expiresAt)
	writeJSON(w, http.StatusOK, loginResponse{OK: true})
}

func (g *Guard) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := sessionTokenFromRequest(r, g.basePath)
	if token != "" {
		_ = g.db.DeleteAuthSession(hashToken(token))
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     cookiePath(g.basePath),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (g *Guard) hasValidSession(r *http.Request) bool {
	token := sessionTokenFromRequest(r, g.basePath)
	if token == "" {
		return false
	}
	session, err := g.db.GetAuthSession(hashToken(token))
	if err != nil {
		return false
	}
	if session.ExpiresAt <= time.Now().Unix() {
		_ = g.db.DeleteAuthSession(session.TokenHash)
		return false
	}
	return true
}

func (g *Guard) isBlocked(clientHash string) (bool, int64) {
	if !g.bruteEnabled {
		return false, 0
	}
	state, err := g.db.GetAuthBruteState(clientHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, 0
		}
		return false, 0
	}
	now := time.Now().Unix()
	if state.BannedUntil > now {
		return true, state.BannedUntil - now
	}
	return false, 0
}

func (g *Guard) recordFailure(clientHash string) {
	if !g.bruteEnabled {
		return
	}
	state, err := g.db.GetAuthBruteState(clientHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		state = db.AuthBruteState{ClientHash: clientHash}
	}
	state.FailCount++
	limit := g.bruteMax
	if state.Trusted {
		limit = g.bruteRelaxedMax
	}
	if state.FailCount >= limit {
		state.BannedUntil = time.Now().Add(g.bruteBan).Unix()
	}
	_ = g.db.UpsertAuthBruteState(state)
}

func (g *Guard) recordSuccess(clientHash string) {
	_ = g.db.MarkAuthClientTrusted(clientHash)
}

func loadOrCreatePepper(database *db.DB) (string, error) {
	raw, err := database.GetSetting(db.AuthPepperSettingKey)
	if err == nil && raw != "" {
		return raw, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	pepper := hex.EncodeToString(buf)
	if err := database.SetSetting(db.AuthPepperSettingKey, pepper); err != nil {
		return "", err
	}
	return pepper, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func sessionTokenFromRequest(r *http.Request, _ string) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return cookie.Value
}

func cookiePath(basePath string) string {
	if basePath == "" || basePath == "/" {
		return "/"
	}
	return basePath
}

func (g *Guard) setSessionCookie(w http.ResponseWriter, token string, expiresAt int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     cookiePath(g.basePath),
		Expires:  time.Unix(expiresAt, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func isPublicPath(route string) bool {
	if route == "/" || route == "/index.html" || route == "/favicon.svg" || route == "/favicon.ico" {
		return true
	}
	if strings.HasPrefix(route, "/assets/") {
		return true
	}
	ext := path.Ext(route)
	switch ext {
	case ".js", ".css", ".svg", ".png", ".jpg", ".jpeg", ".webp", ".woff", ".woff2", ".ttf", ".map", ".wasm":
		return true
	}
	if strings.HasPrefix(route, "/src/") {
		return true
	}
	return false
}

func normalizeBasePath(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return ""
	}
	return "/" + raw
}

func stripBasePath(route, base string) string {
	if base == "" {
		return route
	}
	if route == base {
		return "/"
	}
	if strings.HasPrefix(route, base+"/") {
		return strings.TrimPrefix(route, base)
	}
	return route
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

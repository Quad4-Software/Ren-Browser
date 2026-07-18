// SPDX-License-Identifier: MIT
package config

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"renbrowser/internal/brand"
)

type Runtime struct {
	ReticulumConfig string
	AssetsDir       string
	AssetsZip       string
	ServerHost      string
	ServerPort      int
	TrustProxy      bool
	BasePath        string
	Profile         string
	ExportProfile   string
	ImportProfile   string
	PublicMode      bool
	ResetWindow     bool
	NativeTitlebar  bool
	NoLandlock      bool
	Landlock        bool
	NoSeccomp       bool
	Seccomp         bool
	Auth            bool
	AuthReset       bool
	NoBruteForce    bool
	AuthBruteMax    int
	AuthBruteRelax  int
	AuthBruteBanMin int
	AuthIPWhitelist string
	AuthSessionHrs  int
	LogLevel        string
	Reset           bool
	Headless        bool
	SelfCheck       bool
	Version         bool
}

func ParseFlags() Runtime {
	var cfg Runtime
	flag.BoolVar(&cfg.Version, "v", false, "Show version information and exit")
	flag.BoolVar(&cfg.Version, "version", false, "Show version information and exit")
	flag.StringVar(&cfg.ReticulumConfig, "config", "", "Reticulum config file path")
	flag.StringVar(&cfg.AssetsDir, "assets-dir", "", "Serve frontend from directory instead of embedded assets")
	flag.StringVar(&cfg.AssetsZip, "assets-zip", "", "Serve frontend from zip archive")
	flag.StringVar(&cfg.ServerHost, "host", "", "HTTP server bind host (server mode)")
	flag.IntVar(&cfg.ServerPort, "port", 0, "HTTP server port (server mode)")
	flag.BoolVar(&cfg.TrustProxy, "trust-proxy", false, "Trust X-Forwarded-* headers from reverse proxies")
	flag.StringVar(&cfg.BasePath, "base-path", "", "URL prefix when served behind a reverse proxy subpath")
	flag.StringVar(&cfg.Profile, "profile", "", fmt.Sprintf("Browser profile name (default profile uses ~/%s/%s)", brand.DataDirName, brand.DBFileName))
	flag.StringVar(&cfg.ExportProfile, "export-profile", "", "Export the active profile to a JSON file at startup")
	flag.StringVar(&cfg.ImportProfile, "import-profile", "", "Import profile data from a JSON file at startup")
	flag.BoolVar(&cfg.PublicMode, "public-mode", false, "Store favorites, history, and tabs in the browser (localStorage) instead of the server database")
	flag.BoolVar(&cfg.ResetWindow, "reset-window", false, "Ignore saved window size and position on startup")
	flag.BoolVar(&cfg.Reset, "reset", false, "Reset the browser (delete config, cache, database, and settings)")
	flag.BoolVar(&cfg.NativeTitlebar, "native-titlebar", false, "Use the native OS title bar instead of custom window controls")
	flag.BoolVar(&cfg.NoLandlock, "no-landlock", false, "Disable Landlock LSM filesystem sandbox (Linux)")
	flag.BoolVar(&cfg.Landlock, "landlock", false, "Force Landlock LSM sandbox on Linux even when auto-detection would skip it")
	flag.BoolVar(&cfg.NoSeccomp, "no-seccomp", false, "Disable seccomp-bpf syscall hardening (Linux)")
	flag.BoolVar(&cfg.Seccomp, "seccomp", false, "Force seccomp-bpf hardening on Linux even when auto-detection would skip it")
	flag.BoolVar(&cfg.Auth, "auth", false, "Enable HTTP basic auth for server mode (prompts to set password on first run)")
	flag.BoolVar(&cfg.AuthReset, "auth-reset", false, "Reset server auth password and prompt for a new one")
	flag.BoolVar(&cfg.NoBruteForce, "no-bruteforce-protection", false, "Disable login brute-force IP bans (server mode)")
	flag.IntVar(&cfg.AuthBruteMax, "auth-brute-max", 0, "Failed login attempts before ban when brute-force protection is on (default 3)")
	flag.IntVar(&cfg.AuthBruteRelax, "auth-brute-relaxed-max", 0, "Failed login attempts before ban for trusted clients (default 10)")
	flag.IntVar(&cfg.AuthBruteBanMin, "auth-brute-ban-minutes", 0, "Brute-force ban duration in minutes (default 15)")
	flag.StringVar(&cfg.AuthIPWhitelist, "auth-ip-whitelist", "", "Comma-separated IPs/CIDRs that bypass auth (supports IPv6)")
	flag.IntVar(&cfg.AuthSessionHrs, "auth-session-hours", 0, "Auth session lifetime in hours (default 168)")
	flag.StringVar(&cfg.LogLevel, "log-level", "", "Server log level: debug, info, warn, error")
	flag.BoolVar(&cfg.Headless, "headless", false, "Run in headless mode (exit immediately after startup/self-check)")
	flag.BoolVar(&cfg.SelfCheck, "self-check", false, "Run internal diagnostics and exit with code 0 if healthy, 1 otherwise")
	flag.Parse()
	LoadDotEnv("")
	return ApplyEnv(cfg)
}

func LoadDotEnv(path string) {
	if path == "" {
		path = ".env"
	}
	f, err := os.Open(path) // #nosec G304 -- optional local .env in working directory
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func ApplyEnv(cfg Runtime) Runtime {
	if v := envFirst("REN_BROWSER_CONFIG", "RETICULUM_CONFIG"); v != "" && cfg.ReticulumConfig == "" {
		cfg.ReticulumConfig = v
	}
	if v := os.Getenv("REN_BROWSER_ASSETS_DIR"); v != "" && cfg.AssetsDir == "" {
		cfg.AssetsDir = v
	}
	if v := os.Getenv("REN_BROWSER_ASSETS_ZIP"); v != "" && cfg.AssetsZip == "" {
		cfg.AssetsZip = v
	}
	if v := envFirst("WAILS_SERVER_HOST", "REN_BROWSER_HOST"); v != "" && cfg.ServerHost == "" {
		cfg.ServerHost = v
	}
	if cfg.ServerPort == 0 {
		if v := envFirst("WAILS_SERVER_PORT", "REN_BROWSER_PORT"); v != "" {
			if port, err := strconv.Atoi(v); err == nil {
				cfg.ServerPort = port
			}
		}
	}
	if !cfg.TrustProxy {
		cfg.TrustProxy = envBool("REN_BROWSER_TRUST_PROXY")
	}
	if v := os.Getenv("REN_BROWSER_BASE_PATH"); v != "" && cfg.BasePath == "" {
		cfg.BasePath = v
	}
	if v := os.Getenv("REN_BROWSER_PROFILE"); v != "" && cfg.Profile == "" {
		cfg.Profile = v
	}
	if v := os.Getenv("REN_BROWSER_EXPORT_PROFILE"); v != "" && cfg.ExportProfile == "" {
		cfg.ExportProfile = v
	}
	if v := os.Getenv("REN_BROWSER_IMPORT_PROFILE"); v != "" && cfg.ImportProfile == "" {
		cfg.ImportProfile = v
	}
	if !cfg.PublicMode {
		cfg.PublicMode = envBool("REN_BROWSER_PUBLIC_MODE")
	}
	if !cfg.ResetWindow {
		cfg.ResetWindow = envBool("REN_BROWSER_RESET_WINDOW")
	}
	if !cfg.Reset {
		cfg.Reset = envBool("REN_BROWSER_RESET")
	}
	if !cfg.Headless {
		cfg.Headless = envBool("REN_BROWSER_HEADLESS")
	}
	if !cfg.SelfCheck {
		cfg.SelfCheck = envBool("REN_BROWSER_SELF_CHECK")
	}
	if !cfg.NativeTitlebar {
		cfg.NativeTitlebar = envBool("REN_BROWSER_NATIVE_TITLEBAR")
	}
	if !cfg.Auth {
		cfg.Auth = envBool("REN_BROWSER_AUTH")
	}
	if !cfg.AuthReset {
		cfg.AuthReset = envBool("REN_BROWSER_AUTH_RESET")
	}
	if !cfg.NoBruteForce {
		cfg.NoBruteForce = envBool("REN_BROWSER_NO_BRUTEFORCE_PROTECTION")
	}
	if cfg.AuthBruteMax == 0 {
		cfg.AuthBruteMax = envInt("REN_BROWSER_AUTH_BRUTE_MAX", 3)
	}
	if cfg.AuthBruteRelax == 0 {
		cfg.AuthBruteRelax = envInt("REN_BROWSER_AUTH_BRUTE_RELAXED_MAX", 10)
	}
	if cfg.AuthBruteBanMin == 0 {
		cfg.AuthBruteBanMin = envInt("REN_BROWSER_AUTH_BRUTE_BAN_MINUTES", 15)
	}
	if v := os.Getenv("REN_BROWSER_AUTH_IP_WHITELIST"); v != "" && cfg.AuthIPWhitelist == "" {
		cfg.AuthIPWhitelist = v
	}
	if cfg.AuthSessionHrs == 0 {
		cfg.AuthSessionHrs = envInt("REN_BROWSER_AUTH_SESSION_HOURS", 168)
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = envFirst("REN_BROWSER_LOG_LEVEL", "LOG_LEVEL")
	}
	return cfg
}

func ParseLogLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug", "trace":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func envFirst(keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

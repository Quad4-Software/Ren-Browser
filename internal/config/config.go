// SPDX-License-Identifier: MIT
package config

import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"strings"
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
}

func ParseFlags() Runtime {
	var cfg Runtime
	flag.StringVar(&cfg.ReticulumConfig, "config", "", "Reticulum config file path")
	flag.StringVar(&cfg.AssetsDir, "assets-dir", "", "Serve frontend from directory instead of embedded assets")
	flag.StringVar(&cfg.AssetsZip, "assets-zip", "", "Serve frontend from zip archive")
	flag.StringVar(&cfg.ServerHost, "host", "", "HTTP server bind host (server mode)")
	flag.IntVar(&cfg.ServerPort, "port", 0, "HTTP server port (server mode)")
	flag.BoolVar(&cfg.TrustProxy, "trust-proxy", false, "Trust X-Forwarded-* headers from reverse proxies")
	flag.StringVar(&cfg.BasePath, "base-path", "", "URL prefix when served behind a reverse proxy subpath")
	flag.StringVar(&cfg.Profile, "profile", "", "Browser profile name (default profile uses ~/.renbrowser/renbrowser.db)")
	flag.StringVar(&cfg.ExportProfile, "export-profile", "", "Export the active profile to a JSON file at startup")
	flag.StringVar(&cfg.ImportProfile, "import-profile", "", "Import profile data from a JSON file at startup")
	flag.BoolVar(&cfg.PublicMode, "public-mode", false, "Store favorites, history, and tabs in the browser (localStorage) instead of the server database")
	flag.BoolVar(&cfg.ResetWindow, "reset-window", false, "Ignore saved window size and position on startup")
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
	return cfg
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

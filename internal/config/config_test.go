// SPDX-License-Identifier: MIT
package config_test

import (
	"log/slog"
	"os"
	"testing"

	"renbrowser/internal/config"
)

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("REN_BROWSER_HOST", "0.0.0.0")
	t.Setenv("REN_BROWSER_PORT", "9090")
	t.Setenv("REN_BROWSER_TRUST_PROXY", "true")
	t.Setenv("REN_BROWSER_BASE_PATH", "/ren")
	t.Setenv("REN_BROWSER_AUTH", "true")
	t.Setenv("REN_BROWSER_AUTH_BRUTE_MAX", "5")

	cfg := config.ApplyEnv(config.Runtime{})
	if cfg.ServerHost != "0.0.0.0" {
		t.Fatalf("host = %q", cfg.ServerHost)
	}
	if cfg.ServerPort != 9090 {
		t.Fatalf("port = %d", cfg.ServerPort)
	}
	if !cfg.TrustProxy {
		t.Fatal("expected trust proxy")
	}
	if cfg.BasePath != "/ren" {
		t.Fatalf("base path = %q", cfg.BasePath)
	}
	if !cfg.Auth {
		t.Fatal("expected auth")
	}
	if cfg.AuthBruteMax != 5 {
		t.Fatalf("brute max = %d", cfg.AuthBruteMax)
	}
}

func TestParseLogLevel(t *testing.T) {
	if config.ParseLogLevel("debug") != slog.LevelDebug {
		t.Fatal("expected debug")
	}
	if config.ParseLogLevel("warn") != slog.LevelWarn {
		t.Fatal("expected warn")
	}
	if config.ParseLogLevel("error") != slog.LevelError {
		t.Fatal("expected error")
	}
	if config.ParseLogLevel("") != slog.LevelInfo {
		t.Fatal("empty should default to info")
	}
}

func TestApplyEnvLogLevel(t *testing.T) {
	t.Setenv("REN_BROWSER_LOG_LEVEL", "error")
	cfg := config.ApplyEnv(config.Runtime{})
	if cfg.LogLevel != "error" {
		t.Fatalf("log level = %q", cfg.LogLevel)
	}
}

func TestApplyEnvNativeTitlebar(t *testing.T) {
	t.Setenv("REN_BROWSER_NATIVE_TITLEBAR", "true")
	cfg := config.ApplyEnv(config.Runtime{})
	if !cfg.NativeTitlebar {
		t.Fatal("expected native titlebar")
	}
}

func TestLoadDotEnvDoesNotOverrideExisting(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/.env"
	if err := os.WriteFile(path, []byte("REN_BROWSER_HOST=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REN_BROWSER_HOST", "from-env")
	config.LoadDotEnv(path)
	if got := os.Getenv("REN_BROWSER_HOST"); got != "from-env" {
		t.Fatalf("host = %q", got)
	}
}

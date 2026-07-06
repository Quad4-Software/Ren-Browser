//go:build server

// SPDX-License-Identifier: MIT
package serverauth

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"renbrowser/internal/auth"
	"renbrowser/internal/config"
	"renbrowser/internal/db"
	"renbrowser/internal/serverlog"
)

func Configure(database *db.DB, cfg config.Runtime) (*Guard, error) {
	if database == nil {
		return nil, errors.New("database required")
	}

	enabled, err := database.AuthEnabled()
	if err != nil {
		return nil, err
	}

	if cfg.AuthReset {
		if err := database.ClearAuthCredential(); err != nil {
			return nil, err
		}
		_ = database.ClearAuthSessions()
		_ = database.ClearAuthBruteState()
		enabled = false
		serverlog.Info("auth password reset")
	}

	if cfg.Auth || cfg.AuthReset {
		if !enabled {
			hash, err := promptNewPassword()
			if err != nil {
				return nil, err
			}
			if err := database.SetAuthCredential(hash); err != nil {
				return nil, err
			}
			enabled = true
			serverlog.Info("auth password configured")
		}
	}

	guard, err := NewGuard(database, cfg)
	if err != nil {
		return nil, err
	}
	if enabled {
		guard.Activate()
	}
	return guard, nil
}

func promptNewPassword() (string, error) {
	if !term.IsTerminal(int(syscall.Stdin)) {
		return "", errors.New("auth setup requires an interactive terminal")
	}

	fmt.Fprintln(os.Stderr, "Set a password for Ren Browser server access.")
	password, err := readPassword("New password: ")
	if err != nil {
		return "", err
	}
	confirm, err := readPassword("Confirm password: ")
	if err != nil {
		return "", err
	}
	if password != confirm {
		return "", errors.New("passwords do not match")
	}
	return auth.HashPassword(password)
}

func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	if term.IsTerminal(int(syscall.Stdin)) {
		bytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(bytes)), nil
	}
	defer fmt.Fprintln(os.Stderr)

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

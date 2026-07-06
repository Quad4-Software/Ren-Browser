//go:build !server && !android

// SPDX-License-Identifier: MIT

package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"renbrowser/internal/safe"
)

func screenshotThemeFromEnv() string {
	switch strings.TrimSpace(os.Getenv("REN_BROWSER_SCREENSHOT_THEME")) {
	case "light":
		return "light"
	default:
		return "dark"
	}
}

func screenshotDirFromEnv() (string, bool) {
	raw := strings.TrimSpace(os.Getenv("REN_BROWSER_SCREENSHOT_DIR"))
	if raw == "" {
		return "", false
	}
	for _, r := range raw {
		if r == '/' || r == '-' || r == '_' || r == '.' {
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		log.Printf("screenshot: invalid output directory")
		return "", false
	}
	return filepath.Clean(raw), true
}

func maybeCaptureDesktopScreenshot() {
	dir, ok := screenshotDirFromEnv()
	if !ok {
		return
	}
	theme := screenshotThemeFromEnv()
	root, err := os.Getwd()
	if err != nil {
		log.Printf("screenshot: %v", err)
		return
	}
	script := filepath.Join(root, "build", "scripts", "desktop-screenshot.sh")
	safe.Go("desktop-screenshot", func() {
		time.Sleep(4 * time.Second)
		cmd := exec.Command(script, dir, theme) // #nosec G204 -- argv-only; dir and theme are validated
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("screenshot failed: %v (%q)", err, strings.ReplaceAll(string(out), "\n", " "))
			return
		}
		log.Printf("screenshot saved: %q", strings.TrimSpace(string(out)))
	})
}

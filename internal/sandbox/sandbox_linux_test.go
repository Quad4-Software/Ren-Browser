// SPDX-License-Identifier: MIT

//go:build linux

package sandbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestLandlockFunctional(t *testing.T) {
	if !KernelSupported() {
		t.Skip("Landlock not supported")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	allowedDir := filepath.Join(home, ".renbrowser", "sandbox-test-allowed")
	blockedDir := filepath.Join(home, "sandbox-test-blocked-renbrowser")

	if err := os.MkdirAll(allowedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(blockedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(allowedDir)
		_ = os.RemoveAll(blockedDir)
	})

	allowedFile := filepath.Join(allowedDir, "allowed.txt")
	blockedFile := filepath.Join(blockedDir, "blocked.txt")
	if err := os.WriteFile(allowedFile, []byte("allowed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blockedFile, []byte("blocked"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLandlockHelper$", "-test.v")
	cmd.Env = append(os.Environ(),
		"SANDBOX_LANDLOCK_TEST=1",
		"SANDBOX_ALLOWED_DIR="+allowedDir,
		"SANDBOX_BLOCKED_DIR="+blockedDir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "not supported") {
			t.Skip("Landlock not available in test environment")
		}
		t.Fatalf("helper failed:\n%s", out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("helper did not pass:\n%s", out)
	}
}

func TestLandlockHelper(t *testing.T) {
	if os.Getenv("SANDBOX_LANDLOCK_TEST") != "1" {
		t.Skip("helper subprocess only")
	}

	allowedDir := os.Getenv("SANDBOX_ALLOWED_DIR")
	blockedDir := os.Getenv("SANDBOX_BLOCKED_DIR")
	if allowedDir == "" || blockedDir == "" {
		t.Fatal("missing test dirs")
	}

	opts := Options{
		ForceLandlock: true,
		ServerMode:    true,
		DataDir:       allowedDir,
	}
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		t.Fatalf("PR_SET_NO_NEW_PRIVS: %v", err)
	}
	if err := applyLandlock(opts); err != nil {
		if strings.Contains(err.Error(), "not supported") {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	allowedFile := filepath.Join(allowedDir, "allowed.txt")
	blockedFile := filepath.Join(blockedDir, "blocked.txt")

	if data, err := os.ReadFile(allowedFile); err != nil || string(data) != "allowed" {
		t.Fatalf("allowed read failed: %v data=%q", err, data)
	}
	if _, err := os.ReadFile(blockedFile); err == nil {
		t.Fatal("expected blocked file read to fail")
	}
	if _, err := os.ReadDir(blockedDir); err == nil {
		t.Fatal("expected blocked directory listing to fail")
	}
}

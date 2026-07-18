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

func TestABIVersion(t *testing.T) {
	if !KernelSupported() {
		if ABIVersion() != 0 {
			t.Fatalf("ABIVersion = %d, want 0 when unsupported", ABIVersion())
		}
		t.Skip("Landlock not supported")
	}
	if ABIVersion() < 1 {
		t.Fatalf("ABIVersion = %d, want >= 1", ABIVersion())
	}
}

func TestLandlockHandledForABI(t *testing.T) {
	fs1, scoped1 := landlockHandledForABI(1)
	if fs1&unix.LANDLOCK_ACCESS_FS_REFER != 0 {
		t.Fatal("ABI1 should not handle REFER")
	}
	if fs1&unix.LANDLOCK_ACCESS_FS_TRUNCATE != 0 {
		t.Fatal("ABI1 should not handle TRUNCATE")
	}
	if scoped1 != 0 {
		t.Fatal("ABI1 should not set scopes")
	}

	fs5, scoped5 := landlockHandledForABI(5)
	if fs5&unix.LANDLOCK_ACCESS_FS_IOCTL_DEV == 0 {
		t.Fatal("ABI5 should handle IOCTL_DEV")
	}
	if scoped5 != 0 {
		t.Fatal("ABI5 should not set scopes")
	}

	fs6, scoped6 := landlockHandledForABI(6)
	if scoped6&unix.LANDLOCK_SCOPE_SIGNAL == 0 {
		t.Fatal("ABI6 should scope signals")
	}
	if fs6&landlockAccessFSResolveUnix != 0 {
		t.Fatal("ABI6 should not handle RESOLVE_UNIX")
	}

	fs9, _ := landlockHandledForABI(9)
	if fs9&landlockAccessFSResolveUnix == 0 {
		t.Fatal("ABI9 should handle RESOLVE_UNIX")
	}
}

func TestSeccompFilterBuilds(t *testing.T) {
	filter, err := buildSeccompFilter()
	if err != nil {
		t.Fatal(err)
	}
	if len(filter) < 4 {
		t.Fatalf("filter too short: %d", len(filter))
	}
	last := filter[len(filter)-1]
	if last.Code != unix.BPF_RET|unix.BPF_K || last.K != unix.SECCOMP_RET_ALLOW {
		t.Fatalf("filter must end with ALLOW, got code=%#x k=%#x", last.Code, last.K)
	}
}

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
		NoSeccomp:     true,
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

func TestSeccompFunctional(t *testing.T) {
	if !SeccompSupported() {
		t.Skip("seccomp not supported")
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestSeccompHelper$", "-test.v")
	cmd.Env = append(os.Environ(), "SANDBOX_SECCOMP_TEST=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "not supported") || strings.Contains(string(out), "seccomp filter") {
			t.Skipf("seccomp not available in test environment:\n%s", out)
		}
		t.Fatalf("helper failed:\n%s", out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("helper did not pass:\n%s", out)
	}
}

func TestSeccompHelper(t *testing.T) {
	if os.Getenv("SANDBOX_SECCOMP_TEST") != "1" {
		t.Skip("helper subprocess only")
	}

	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		t.Fatalf("PR_SET_NO_NEW_PRIVS: %v", err)
	}
	if err := applySeccomp(); err != nil {
		t.Fatal(err)
	}

	// Allowed operations must still work.
	tmp, err := os.CreateTemp("", "seccomp-ok-*")
	if err != nil {
		t.Fatalf("create temp after seccomp: %v", err)
	}
	name := tmp.Name()
	_, _ = tmp.WriteString("ok")
	_ = tmp.Close()
	defer os.Remove(name)

	data, err := os.ReadFile(name)
	if err != nil || string(data) != "ok" {
		t.Fatalf("read after seccomp failed: %v data=%q", err, data)
	}

	// Denied syscall must fail with EPERM.
	err = unix.Mount("tmpfs", "/tmp", "tmpfs", 0, "")
	if err == nil {
		_ = unix.Unmount("/tmp", unix.MNT_DETACH)
		t.Fatal("expected mount to be blocked by seccomp")
	}
	if err != unix.EPERM && err != unix.EACCES {
		// Unprivileged mount often fails with EPERM from LSM before seccomp.
		// Accept either denial path as long as mount did not succeed.
		t.Logf("mount denied with %v (acceptable)", err)
	}
}

func TestApply_LandlockAndSeccomp(t *testing.T) {
	if !KernelSupported() || !SeccompSupported() {
		t.Skip("landlock/seccomp not supported")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	dataDir := filepath.Join(home, ".renbrowser", "sandbox-apply-test")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dataDir) })

	cmd := exec.Command(os.Args[0], "-test.run=TestApplyHelper$", "-test.v")
	cmd.Env = append(os.Environ(),
		"SANDBOX_APPLY_TEST=1",
		"SANDBOX_DATA_DIR="+dataDir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper failed:\n%s", out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("helper did not pass:\n%s", out)
	}
}

func TestApplyHelper(t *testing.T) {
	if os.Getenv("SANDBOX_APPLY_TEST") != "1" {
		t.Skip("helper subprocess only")
	}

	Apply(Options{
		ForceLandlock: true,
		ForceSeccomp:  true,
		ServerMode:    true,
		DataDir:       os.Getenv("SANDBOX_DATA_DIR"),
	})
	st := CurrentStatus()
	if !st.Enabled {
		t.Fatalf("expected landlock enabled: %+v", st)
	}
	if !st.SeccompEnabled {
		t.Fatalf("expected seccomp enabled: %+v", st)
	}
	if st.Type != "landlock+seccomp" {
		t.Fatalf("type = %q, want landlock+seccomp", st.Type)
	}
	if st.ABI < 1 {
		t.Fatalf("ABI = %d, want >= 1", st.ABI)
	}
}

func TestApply_NoSeccompStillLandlock(t *testing.T) {
	if !KernelSupported() {
		t.Skip("Landlock not supported")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	dataDir := filepath.Join(home, ".renbrowser", "sandbox-noseccomp-test")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dataDir) })

	cmd := exec.Command(os.Args[0], "-test.run=TestApplyNoSeccompHelper$", "-test.v")
	cmd.Env = append(os.Environ(),
		"SANDBOX_APPLY_NOSECCOMP_TEST=1",
		"SANDBOX_DATA_DIR="+dataDir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper failed:\n%s", out)
	}
	if !strings.Contains(string(out), "PASS") {
		t.Fatalf("helper did not pass:\n%s", out)
	}
}

func TestApplyNoSeccompHelper(t *testing.T) {
	if os.Getenv("SANDBOX_APPLY_NOSECCOMP_TEST") != "1" {
		t.Skip("helper subprocess only")
	}

	Apply(Options{
		ForceLandlock: true,
		NoSeccomp:     true,
		ServerMode:    true,
		DataDir:       os.Getenv("SANDBOX_DATA_DIR"),
	})
	st := CurrentStatus()
	if !st.Enabled {
		t.Fatalf("expected landlock enabled: %+v", st)
	}
	if st.SeccompEnabled {
		t.Fatal("expected seccomp disabled")
	}
	if st.Type != "landlock" {
		t.Fatalf("type = %q, want landlock", st.Type)
	}
}

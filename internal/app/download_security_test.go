// SPDX-License-Identifier: MIT
package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDownloadAbsoluteNameEscapesJoinDir(t *testing.T) {
	dir := t.TempDir()
	escaped := filepath.Join(dir, "/tmp/renbrowser-download-escape-probe")
	t.Logf("Join(%q, %q) -> %q", dir, "/tmp/renbrowser-download-escape-probe", escaped)

	safe := sanitizeDownloadFilename("/tmp/evil.bin")
	if strings.Contains(safe, string(filepath.Separator)) {
		t.Fatalf("sanitize left separator: %q", safe)
	}
	if safe != "evil.bin" {
		t.Fatalf("sanitize got %q want evil.bin", safe)
	}

	// Desktop writeDownloadBytes trusts name. With sanitize, write stays in dir.
	dest, err := writeDownloadBytes(dir, safe, []byte("ok"))
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(dir, dest)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("wrote outside dir: dest=%q rel=%q", dest, rel)
	}
	_ = os.Remove(dest)

	// Defense gap: unsanitized absolute name may leave dir depending on Join.
	if filepath.IsAbs(escaped) && escaped != filepath.Join(dir, "tmp", "renbrowser-download-escape-probe") {
		t.Logf("unsanitized absolute name would write outside download dir via Join")
	}
}

func TestDownloadCancelOverwrittenBySetStatus(t *testing.T) {
	mgr := newDownloadManager()
	id := mgr.start("deadbeef:/file/x.bin", "x.bin")
	if !mgr.cancel(id) {
		t.Fatal("cancel failed")
	}
	item, ok := mgr.findByID(id)
	if !ok || item.Status != DownloadStatusCanceled {
		t.Fatalf("status=%v ok=%v", item.Status, ok)
	}
	mgr.setStatus(id, DownloadStatusRetrying)
	item, ok = mgr.findByID(id)
	if !ok {
		t.Fatal("missing item")
	}
	if item.Status != DownloadStatusCanceled {
		t.Fatalf("Canceled must stick across setStatus(Retrying), got %v", item.Status)
	}
}

func TestDownloadWriteStaysUnderDirWhenSanitized(t *testing.T) {
	dir := t.TempDir()
	name := sanitizeDownloadFilename(`../../etc/passwd`)
	dest, err := writeDownloadBytes(dir, name, []byte("ok"))
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(dir, dest)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("wrote outside download dir: dest=%q rel=%q err=%v", dest, rel, err)
	}
}

func TestDownloadManagerTracksManyEntries(t *testing.T) {
	// Tracking many entries is fine. Concurrent transfer slots live on BrowserService.
	mgr := newDownloadManager()
	start := time.Now()
	for range 32 {
		_ = mgr.start("deadbeef:/file/blob.bin", "blob.bin")
	}
	if time.Since(start) > time.Second {
		t.Fatal("manager start unexpectedly slow")
	}
	if n := mgr.runningCount(); n < 32 {
		t.Fatalf("runningCount=%d want >=32", n)
	}
}

func TestDownloadFilenameWindowsReservedAndTrailingNames(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{"CON", "_CON"},
		{"con.txt", "_con.txt"},
		{"NUL", "_NUL"},
		{"COM1", "_COM1"},
		{"lpt9.bin", "_lpt9.bin"},
		{"name.", "name"},
		{"name   ", "name"},
	} {
		if got := sanitizeDownloadFilename(tc.in); got != tc.want {
			t.Fatalf("sanitize(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestDownloadWriteDoesNotFollowExistingSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires extra privileges on Windows")
	}
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "x.bin")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	dest, err := writeDownloadBytes(dir, "x.bin", []byte("new"))
	if err != nil {
		t.Fatal(err)
	}
	if dest == link {
		t.Fatal("write reused existing symlink path")
	}
	data, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "keep" {
		t.Fatalf("symlink target overwritten: %q", data)
	}
}

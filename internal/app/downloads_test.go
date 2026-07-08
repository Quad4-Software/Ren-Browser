// SPDX-License-Identifier: MIT
package app

import (
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/rns"
)

func TestDownloadNameFromURL(t *testing.T) {
	got := downloadNameFromURL("abb3ebcd03cb2388a838e70c001291f9:/file/guide.zip?token=abc")
	if got != "guide.zip" {
		t.Fatalf("got %q", got)
	}
	if downloadNameFromURL("config:") != "reticulum.conf" {
		t.Fatalf("config download name = %q", downloadNameFromURL("config:"))
	}
}

func TestSanitizeDownloadFilename(t *testing.T) {
	got := sanitizeDownloadFilename(`../evil/name?.bin`)
	if got != "name_.bin" {
		t.Fatalf("got %q", got)
	}
}

func TestIsRootLevelDownloadDir(t *testing.T) {
	if !isRootLevelDownloadDir("/Downloads") {
		t.Fatal("expected /Downloads to be rejected")
	}
	if isRootLevelDownloadDir("/home/user/Downloads") {
		t.Fatal("expected normal path to be accepted")
	}
}

func TestIOSDownloadDirNeedsReset(t *testing.T) {
	root := "/private/var/mobile/Containers/Data/Application/UUID/Documents"
	cases := []struct {
		dir  string
		want bool
	}{
		{"", true},
		{"/Downloads", true},
		{"/private/var/mobile/Containers/Data/Application/UUID/Downloads", true},
		{root + "/Downloads", false},
		{filepath.Join(root, "Downloads", "subdir"), false},
		{os.TempDir(), true},
	}
	for _, tc := range cases {
		if got := iosDownloadDirNeedsReset(tc.dir, root); got != tc.want {
			t.Fatalf("iosDownloadDirNeedsReset(%q) = %v; want %v", tc.dir, got, tc.want)
		}
	}
}

func TestIsTempDownloadDir(t *testing.T) {
	temp := filepath.Clean(os.TempDir())
	if !isTempDownloadDir(temp) {
		t.Fatal("expected temp dir to be rejected")
	}
	nested := filepath.Join(temp, "renbrowser")
	if !isTempDownloadDir(nested) {
		t.Fatal("expected nested temp dir to be rejected")
	}
}

func TestGetDownloadDirDefaultsAndPersists(t *testing.T) {
	svc := newTestBrowserService(t)
	dir := svc.GetDownloadDir()
	if isTempDownloadDir(dir) {
		t.Fatalf("default download dir must not be temp: %q", dir)
	}
	if dir != svc.GetDownloadDir() {
		t.Fatal("expected persisted download dir on second read")
	}
}

func TestUniqueFilePath(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "page.mu")
	if err := writeTestFile(base); err != nil {
		t.Fatal(err)
	}
	next := uniqueFilePath(base)
	if next == base {
		t.Fatal("expected unique path")
	}
}

func TestClearDownloadHistory(t *testing.T) {
	svc := newTestBrowserService(t)
	if _, err := svc.SaveTextToDownloadDir("page.mu", "hello"); err != nil {
		t.Fatal(err)
	}
	if len(svc.ListDownloads()) != 1 {
		t.Fatal("expected one download in history")
	}
	result := svc.ClearDownloadHistory()
	if !result.OK {
		t.Fatalf("result=%#v", result)
	}
	if len(svc.ListDownloads()) != 0 {
		t.Fatal("expected history cleared")
	}
}

func TestListDownloads(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(home, "renbrowser-dl-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	root := t.TempDir()
	dbPath := filepath.Join(root, "profile.db")

	stack, err := rns.NewStack(filepath.Join(root, "config"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stack.Stop() })
	svc, err := NewBrowserServiceWithOptions(stack, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = svc.Store().Close() })
	svc.SetDownloadDir(dir)

	if _, err := svc.SaveTextToDownloadDir("page.mu", "hello"); err != nil {
		t.Fatal(err)
	}

	items := svc.ListDownloads()
	if len(items) != 1 {
		t.Fatalf("len = %d", len(items))
	}
	if items[0].Name != "page.mu" {
		t.Fatalf("name = %q", items[0].Name)
	}
}

func writeTestFile(path string) error {
	return os.WriteFile(path, []byte("x"), 0o600)
}

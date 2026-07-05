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
	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewBrowserService(stack, nil)
	if err != nil {
		t.Fatal(err)
	}
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

	dbPath := filepath.Join(t.TempDir(), "profile.db")

	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewBrowserServiceWithOptions(stack, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
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

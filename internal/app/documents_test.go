// SPDX-License-Identifier: MIT
package app

import (
	"os"
	"path/filepath"
	"testing"
)

func testDownloadDir(t *testing.T) string {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(home, "renbrowser-doc-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func TestDocumentPage(t *testing.T) {
	dir := testDownloadDir(t)
	pdfPath := filepath.Join(dir, "sample.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 test"), 0o600); err != nil {
		t.Fatal(err)
	}

	svc := newTestBrowserService(t)
	svc.SetDownloadDir(dir)

	rawURL := documentURL(pdfPath, dir)
	page := svc.documentPage(rawURL, false)
	if page.Error != "" {
		t.Fatalf("unexpected error: %s", page.Error)
	}
	if page.ContentType != "pdf" {
		t.Fatalf("content type = %q", page.ContentType)
	}
	if page.BinaryB64 == "" {
		t.Fatal("expected binary payload")
	}
	if page.HTML != "" {
		t.Fatalf("html should be empty, got %q", page.HTML)
	}
}

func TestDocumentPageRejectsOutsideDownloadDir(t *testing.T) {
	dir := testDownloadDir(t)
	outside := filepath.Join(t.TempDir(), "other.pdf")
	if err := os.WriteFile(outside, []byte("%PDF-1.4"), 0o600); err != nil {
		t.Fatal(err)
	}

	svc := newTestBrowserService(t)
	svc.SetDownloadDir(dir)

	page := svc.documentPage(documentURL(outside, dir), false)
	if page.Error == "" {
		t.Fatal("expected error for path outside download dir")
	}
}

func TestDocumentPageTooLarge(t *testing.T) {
	t.Setenv("REN_BROWSER_MAX_PAGE_BYTES", "8")
	dir := testDownloadDir(t)
	pdfPath := filepath.Join(dir, "big.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 big"), 0o600); err != nil {
		t.Fatal(err)
	}

	svc := newTestBrowserService(t)
	svc.SetDownloadDir(dir)
	page := svc.documentPage(documentURL(pdfPath, dir), false)
	if page.Error == "" {
		t.Fatal("expected too-large error")
	}
}

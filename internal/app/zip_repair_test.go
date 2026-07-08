// SPDX-License-Identifier: MIT
package app

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestRepairZipIfNeeded(t *testing.T) {
	body := stripZipCentralDirectory(t, loadFixtureEpub(t))
	repaired, err := repairZipIfNeeded(body)
	if err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if len(repaired) <= len(body) {
		t.Fatal("expected repaired zip to grow")
	}
	reader, err := zip.NewReader(bytes.NewReader(repaired), int64(len(repaired)))
	if err != nil {
		t.Fatalf("repaired zip unreadable: %v", err)
	}
	if len(reader.File) == 0 {
		t.Fatal("expected repaired entries")
	}

	again, err := repairZipIfNeeded(repaired)
	if err != nil {
		t.Fatalf("second repair failed: %v", err)
	}
	if len(again) != len(repaired) {
		t.Fatalf("expected stable repair, got %d vs %d", len(again), len(repaired))
	}
}

func loadFixtureEpub(t *testing.T) []byte {
	t.Helper()
	sample := "/home/user1/Downloads/Hey Nostradamus! (2009) - Douglas Coupland.epub"
	raw, err := os.ReadFile(sample)
	if err != nil {
		t.Skip("sample epub not available")
	}
	return raw
}

func stripZipCentralDirectory(t *testing.T, body []byte) []byte {
	t.Helper()
	searchStart := len(body) - zipEndCentralSize
	minStart := len(body) - 65557
	if minStart < 0 {
		minStart = 0
	}
	var eocdOffset int = -1
	for i := searchStart; i >= minStart; i-- {
		if binary.LittleEndian.Uint32(body[i:]) == zipEndCentralSig {
			eocdOffset = i
			break
		}
	}
	if eocdOffset < 0 {
		t.Fatal("fixture epub has no EOCD")
	}
	cdOffset := int(binary.LittleEndian.Uint32(body[eocdOffset+16:]))
	return append([]byte(nil), body[:cdOffset]...)
}

func TestRepairZipIfNeededRejectsInvalidHeader(t *testing.T) {
	if _, err := repairZipIfNeeded([]byte("not a zip")); err == nil {
		t.Fatal("expected invalid header error")
	}
}

func TestDocumentPageOpensCouplandEpub(t *testing.T) {
	sample := "/home/user1/Downloads/Hey Nostradamus! (2009) - Douglas Coupland.epub"
	if _, err := os.Stat(sample); err != nil {
		t.Skip("sample epub not available")
	}
	dir := testDownloadDir(t)
	dest := filepath.Join(dir, "coupland.epub")
	raw, err := os.ReadFile(sample)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	svc := newTestBrowserService(t)
	svc.SetDownloadDir(dir)
	page := svc.documentPage(documentURL(dest, dir), false)
	if page.Error != "" {
		t.Fatalf("unexpected error: %s", page.Error)
	}
	if page.ContentType != "epub" {
		t.Fatalf("content type = %q", page.ContentType)
	}
	if page.BinaryB64 == "" {
		t.Fatal("expected binary payload")
	}
}

func TestDocumentPageRepairsEpubWithoutCentralDirectory(t *testing.T) {
	sample := "/home/user1/Downloads/23 Years on Fire (2014) - Joel Shepherd.epub"
	if _, err := os.Stat(sample); err != nil {
		t.Skip("sample epub not available")
	}
	dir := testDownloadDir(t)
	dest := filepath.Join(dir, "joel.epub")
	raw, err := os.ReadFile(sample)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	svc := newTestBrowserService(t)
	svc.SetDownloadDir(dir)
	page := svc.documentPage(documentURL(dest, dir), false)
	if page.Error != "" {
		t.Fatalf("unexpected error: %s", page.Error)
	}
	if page.ContentType != "epub" {
		t.Fatalf("content type = %q", page.ContentType)
	}
	if page.BinaryB64 == "" {
		t.Fatal("expected binary payload")
	}
}

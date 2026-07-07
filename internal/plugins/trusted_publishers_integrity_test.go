// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/db"
	"renbrowser/internal/paths"
)

func TestTrustedPublishersIntegrityDetectsExternalEdit(t *testing.T) {
	tmp := t.TempDir()
	paths.SetDataRoot(tmp)
	t.Cleanup(func() { paths.SetDataRoot("") })

	database, err := db.Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := InitTrustedPublishersIntegrity(database); err != nil {
		t.Fatal(err)
	}
	if err := AddUserTrustedPublisherWithStore(database, "abc123", "Publisher"); err != nil {
		t.Fatal(err)
	}

	path := userTrustedPublishersPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tamperedRaw := []byte(`{"publishers":[{"identity":"abc123","name":"Evil Publisher"}]}`)
	if err := os.WriteFile(path, tamperedRaw, 0o600); err != nil {
		t.Fatal(err)
	}
	_ = raw

	if err := VerifyUserTrustedPublishersIntegrity(database); err != nil {
		t.Fatal(err)
	}
	tampered, reason := UserTrustedPublishersTampered()
	if !tampered || reason == "" {
		t.Fatalf("expected tamper detection, tampered=%v reason=%q", tampered, reason)
	}
	publishers := loadUserTrustedPublishers()
	if len(publishers) != 0 {
		t.Fatalf("expected user publishers to be ignored when tampered, got %d", len(publishers))
	}
}

func TestTrustedPublishersIntegrityResetsAfterLegitimateUpdate(t *testing.T) {
	tmp := t.TempDir()
	paths.SetDataRoot(tmp)
	t.Cleanup(func() { paths.SetDataRoot("") })

	database, err := db.Open(filepath.Join(tmp, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := InitTrustedPublishersIntegrity(database); err != nil {
		t.Fatal(err)
	}
	if err := AddUserTrustedPublisherWithStore(database, "abc123", "Publisher"); err != nil {
		t.Fatal(err)
	}
	if err := AddUserTrustedPublisherWithStore(database, "def456", "Other"); err != nil {
		t.Fatal(err)
	}
	tampered, _ := UserTrustedPublishersTampered()
	if tampered {
		t.Fatal("expected trusted publishers to remain valid after managed update")
	}
}

// SPDX-License-Identifier: MIT
package rns

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"quad4/reticulum-go/pkg/identity"
)

func TestOpenIdentityRegistryCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")

	reg, err := OpenIdentityRegistry(storage)
	if err != nil {
		t.Fatal(err)
	}
	items := reg.List()
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}
	if !items[0].Active {
		t.Fatal("expected first identity to be active")
	}
	if items[0].Hash == "" {
		t.Fatal("expected hash")
	}
	if _, err := os.Stat(transportIdentityPath(storage)); err != nil {
		t.Fatalf("transport_identity missing: %v", err)
	}
}

func TestOpenIdentityRegistryRejectsEmptyStorageDir(t *testing.T) {
	if _, err := OpenIdentityRegistry(""); !errors.Is(err, ErrStorageDirEmpty) {
		t.Fatalf("err = %v, want ErrStorageDirEmpty", err)
	}
}

func TestOpenIdentityRegistryRejectsCorruptRegistry(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")
	if err := os.MkdirAll(storage, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(identityRegistryPath(storage), []byte("{not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := OpenIdentityRegistry(storage)
	if !errors.Is(err, ErrRegistryCorrupt) {
		t.Fatalf("err = %v, want ErrRegistryCorrupt", err)
	}
}

func TestOpenIdentityRegistryRejectsEmptyRegistryFile(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")
	if err := os.MkdirAll(storage, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(identityRegistryPath(storage), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := OpenIdentityRegistry(storage)
	if !errors.Is(err, ErrRegistryCorrupt) {
		t.Fatalf("err = %v, want ErrRegistryCorrupt", err)
	}
}

func TestMigrateLegacyTransportIdentity(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")
	if err := os.MkdirAll(storage, 0o700); err != nil {
		t.Fatal(err)
	}
	legacy, err := identity.New()
	if err != nil {
		t.Fatal(err)
	}
	if err := legacy.ToFile(transportIdentityPath(storage)); err != nil {
		t.Fatal(err)
	}

	reg, err := OpenIdentityRegistry(storage)
	if err != nil {
		t.Fatal(err)
	}
	items := reg.List()
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}
	if items[0].Name != "Default" {
		t.Fatalf("name = %q, want Default", items[0].Name)
	}
	if items[0].Hash != legacy.GetHexHash() {
		t.Fatalf("hash = %q, want %q", items[0].Hash, legacy.GetHexHash())
	}
}

func TestIdentityRegistryMultipleAndSwitch(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")

	reg, err := OpenIdentityRegistry(storage)
	if err != nil {
		t.Fatal(err)
	}
	first, err := reg.Create("Work")
	if err != nil {
		t.Fatal(err)
	}
	second, err := reg.Create("Personal")
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.List()) != 3 {
		t.Fatalf("len = %d, want 3", len(reg.List()))
	}
	ident, err := reg.SetActive(second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ident.GetHexHash() != second.Hash {
		t.Fatalf("active hash = %q, want %q", ident.GetHexHash(), second.Hash)
	}
	active, err := reg.ActiveRecord()
	if err != nil {
		t.Fatal(err)
	}
	if !active.Active || active.ID != second.ID {
		t.Fatalf("active = %+v", active)
	}
	if _, err := reg.SetActive(second.ID); !errors.Is(err, ErrIdentityAlreadyActive) {
		t.Fatalf("SetActive same id err = %v", err)
	}
	if err := reg.Delete(first.ID); err != nil {
		t.Fatal(err)
	}
	if len(reg.List()) != 2 {
		t.Fatalf("len after delete = %d, want 2", len(reg.List()))
	}
	if err := reg.Delete(second.ID); !errors.Is(err, ErrCannotDeleteActive) {
		t.Fatalf("delete active err = %v", err)
	}
}

func TestIdentityRegistryImportExport(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")
	exportPath := filepath.Join(dir, "exported_identity")

	reg, err := OpenIdentityRegistry(storage)
	if err != nil {
		t.Fatal(err)
	}
	created, err := reg.Create("Portable")
	if err != nil {
		t.Fatal(err)
	}
	if err := reg.Export(created.ID, exportPath); err != nil {
		t.Fatal(err)
	}
	_, err = reg.ImportFromFile(exportPath, "Duplicate")
	if !errors.Is(err, ErrIdentityDuplicate) {
		t.Fatalf("duplicate import in same registry err = %v", err)
	}

	otherStorage := filepath.Join(t.TempDir(), "storage")
	otherReg, err := OpenIdentityRegistry(otherStorage)
	if err != nil {
		t.Fatal(err)
	}
	imported, err := otherReg.ImportFromFile(exportPath, "Imported")
	if err != nil {
		t.Fatal(err)
	}
	if imported.Hash != created.Hash {
		t.Fatalf("hash = %q, want %q", imported.Hash, created.Hash)
	}
}

func TestIdentityRegistryRename(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")

	reg, err := OpenIdentityRegistry(storage)
	if err != nil {
		t.Fatal(err)
	}
	items := reg.List()
	renamed, err := reg.Rename(items[0].ID, "Primary")
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Name != "Primary" {
		t.Fatalf("name = %q", renamed.Name)
	}
	_, err = reg.Rename(items[0].ID, "   ")
	if !errors.Is(err, ErrIdentityNameEmpty) {
		t.Fatalf("rename empty err = %v", err)
	}
}

func TestIdentityRegistryValidationErrors(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")
	reg, err := OpenIdentityRegistry(storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = reg.Create("   ")
	if !errors.Is(err, ErrIdentityNameEmpty) {
		t.Fatalf("create empty name err = %v", err)
	}
	_, err = reg.Create(strings.Repeat("a", maxIdentityNameLen+1))
	if !errors.Is(err, ErrIdentityNameTooLong) {
		t.Fatalf("create long name err = %v", err)
	}
	_, err = reg.Rename(reg.List()[0].ID, strings.Repeat("b", maxIdentityNameLen+1))
	if !errors.Is(err, ErrIdentityNameTooLong) {
		t.Fatalf("rename long name err = %v", err)
	}
	if err := reg.Delete("../escape"); !errors.Is(err, ErrIdentityIDInvalid) {
		t.Fatalf("delete traversal err = %v", err)
	}
	if err := reg.Export(reg.List()[0].ID, ""); err == nil {
		t.Fatal("expected export path error")
	}
	_, err = reg.ImportFromFile("", "Imported")
	if !errors.Is(err, ErrInvalidIdentityFile) {
		t.Fatalf("import empty path err = %v", err)
	}
}

func TestIdentityRegistryRejectsPathTraversalID(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")
	reg, err := OpenIdentityRegistry(storage)
	if err != nil {
		t.Fatal(err)
	}
	_, err = reg.SetActive("../" + strings.Repeat("a", 30))
	if !errors.Is(err, ErrIdentityIDInvalid) {
		t.Fatalf("SetActive traversal err = %v", err)
	}
}

func TestIdentityRegistryCorruptDuplicateID(t *testing.T) {
	dir := t.TempDir()
	storage := filepath.Join(dir, "storage")
	if err := os.MkdirAll(identitiesDir(storage), 0o700); err != nil {
		t.Fatal(err)
	}
	dupID := strings.Repeat("a", identityIDLen)
	payload := identityRegistryFile{
		Version:  identityRegistryVersion,
		ActiveID: dupID,
		Items: []identityRegistryItem{
			{ID: dupID, Name: "One", Hash: "hash1", CreatedAt: 1},
			{ID: dupID, Name: "Two", Hash: "hash2", CreatedAt: 2},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(identityRegistryPath(storage), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = OpenIdentityRegistry(storage)
	if !errors.Is(err, ErrRegistryCorrupt) {
		t.Fatalf("err = %v, want ErrRegistryCorrupt", err)
	}
}

func TestStackSwitchIdentity(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config")
	stack, err := NewStack(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stack.Stop() })

	reg := stack.Identities()
	created, err := reg.Create("Alt")
	if err != nil {
		t.Fatal(err)
	}
	if err := stack.SwitchIdentity(created.ID); err != nil {
		t.Fatal(err)
	}
	if stack.IdentityHash() != created.Hash {
		t.Fatalf("hash = %q, want %q", stack.IdentityHash(), created.Hash)
	}
	if err := stack.SwitchIdentity(""); !errors.Is(err, ErrIdentityIDInvalid) {
		t.Fatalf("empty id err = %v", err)
	}
}

func TestValidateIdentityID(t *testing.T) {
	valid, err := newIdentityID()
	if err != nil {
		t.Fatal(err)
	}
	if err := validateIdentityID(valid); err != nil {
		t.Fatalf("valid id rejected: %v", err)
	}
	cases := []string{"", "short", strings.Repeat("g", identityIDLen), "../" + strings.Repeat("a", 29)}
	for _, id := range cases {
		if err := validateIdentityID(id); !errors.Is(err, ErrIdentityIDInvalid) {
			t.Fatalf("id %q: err = %v", id, err)
		}
	}
}

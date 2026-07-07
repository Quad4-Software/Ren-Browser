// SPDX-License-Identifier: MIT
package rns

import (
	"os"
	"path/filepath"
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
	if err := reg.Delete(first.ID); err != nil {
		t.Fatal(err)
	}
	if len(reg.List()) != 2 {
		t.Fatalf("len after delete = %d, want 2", len(reg.List()))
	}
	if err := reg.Delete(second.ID); err == nil {
		t.Fatal("expected error deleting active identity")
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
	imported, err := reg.ImportFromFile(exportPath, "Imported")
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
}

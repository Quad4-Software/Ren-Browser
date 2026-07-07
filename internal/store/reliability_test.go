// SPDX-License-Identifier: MIT
package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/db"
	"renbrowser/internal/store"
)

func TestStoreHealthReportsCorruptDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.db")
	if err := os.WriteFile(path, []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	health := st.Health()
	if health.OK {
		t.Fatal("expected corrupt health")
	}
	if health.Kind != "database_corrupt" {
		t.Fatalf("kind=%q", health.Kind)
	}
}

func TestStoreRecoversFromBackupOnCorruptOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recover.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	st.SaveTabs([]store.TabSnapshot{{ID: "tab", Title: "Saved", URL: "deadbeef:/page/saved.mu"}})
	backupPath, err := st.Backup()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}

	st2, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()

	health := st2.Health()
	if !health.OK {
		t.Fatalf("health=%#v; expected recovery from %s", health, backupPath)
	}
	tabs := st2.Tabs()
	if len(tabs) != 1 || tabs[0].Title != "Saved" {
		t.Fatalf("tabs=%#v", tabs)
	}
}

func TestStoreResetCreatesBackup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reset.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	st.SaveTabs([]store.TabSnapshot{{ID: "tab", Title: "Before Reset", URL: "deadbeef:/page/before.mu"}})
	if err := st.Reset(); err != nil {
		t.Fatal(err)
	}
	if len(st.Tabs()) != 0 {
		t.Fatal("expected empty tabs after reset")
	}
	backup, err := db.LatestBackup(path)
	if err != nil {
		t.Fatal(err)
	}
	if backup == "" {
		t.Fatal("expected backup file before reset")
	}
}

func TestStoreSaveTabsSurvivesUncleanClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unclean.db")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	st.SaveTabs([]store.TabSnapshot{
		{ID: "a", Title: "A", URL: "deadbeef:/page/a.mu", Active: true},
	})
	if err := st.DB().Close(); err != nil {
		t.Fatal(err)
	}

	st2, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()

	tabs := st2.Tabs()
	if len(tabs) != 1 || tabs[0].Title != "A" {
		t.Fatalf("tabs=%#v", tabs)
	}
}

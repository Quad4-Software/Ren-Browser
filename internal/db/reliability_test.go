// SPDX-License-Identifier: MIT
package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJournalModeIsWAL(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "wal.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	mode, err := database.PragmaString("journal_mode")
	if err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode=%q; want wal", mode)
	}
}

func TestDurabilityPragmasConfigured(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "pragma.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	timeout, err := database.PragmaInt("busy_timeout")
	if err != nil {
		t.Fatal(err)
	}
	if timeout != 5000 {
		t.Fatalf("busy_timeout=%d; want 5000", timeout)
	}

	foreignKeys, err := database.PragmaInt("foreign_keys")
	if err != nil {
		t.Fatal(err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys=%d; want 1", foreignKeys)
	}

	syncMode, err := database.PragmaInt("synchronous")
	if err != nil {
		t.Fatal(err)
	}
	if syncMode != 2 {
		t.Fatalf("synchronous=%d; want 2 (FULL)", syncMode)
	}

	version, err := database.SchemaVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version != schemaVersion {
		t.Fatalf("user_version=%d; want %d", version, schemaVersion)
	}
}

func TestSaveTabsSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tabs.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	initial := []TabRow{
		{ID: "a", Title: "A", URL: "deadbeef:/page/a.mu", Active: true},
		{ID: "b", Title: "B", URL: "deadbeef:/page/b.mu"},
	}
	if err := database.SaveTabs(initial); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	database, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	tabs, err := database.Tabs()
	if err != nil {
		t.Fatal(err)
	}
	if len(tabs) != 2 || tabs[0].ID != "a" || tabs[1].ID != "b" {
		t.Fatalf("tabs=%#v", tabs)
	}
}

func TestSaveTabsRollbackPreservesExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rollback.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	seed := []TabRow{{ID: "keep", Title: "Keep", URL: "deadbeef:/page/keep.mu"}}
	if err := database.SaveTabs(seed); err != nil {
		t.Fatal(err)
	}

	saveTabsAfterInsertHook = func(inserted int) error {
		if inserted == 2 {
			return errSimulatedWriteFailure
		}
		return nil
	}
	defer func() { saveTabsAfterInsertHook = nil }()

	err = database.SaveTabs([]TabRow{
		{ID: "keep", Title: "Changed", URL: "deadbeef:/page/changed.mu"},
		{ID: "new", Title: "New", URL: "deadbeef:/page/new.mu"},
	})
	if err == nil {
		t.Fatal("expected simulated write failure")
	}

	tabs, err := database.Tabs()
	if err != nil {
		t.Fatal(err)
	}
	if len(tabs) != 1 {
		t.Fatalf("tabs=%#v; want single preserved tab", tabs)
	}
	if tabs[0].Title != "Keep" {
		t.Fatalf("title=%q; want preserved value", tabs[0].Title)
	}
}

func TestSaveTabsUpsertAndPrune(t *testing.T) {
	path := filepath.Join(t.TempDir(), "upsert.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err := database.SaveTabs([]TabRow{
		{ID: "a", Title: "A", URL: "deadbeef:/page/a.mu"},
		{ID: "b", Title: "B", URL: "deadbeef:/page/b.mu"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := database.SaveTabs([]TabRow{
		{ID: "a", Title: "A2", URL: "deadbeef:/page/a2.mu"},
		{ID: "c", Title: "C", URL: "deadbeef:/page/c.mu"},
	}); err != nil {
		t.Fatal(err)
	}

	tabs, err := database.Tabs()
	if err != nil {
		t.Fatal(err)
	}
	if len(tabs) != 2 {
		t.Fatalf("tabs=%#v", tabs)
	}
	if tabs[0].Title != "A2" || tabs[1].ID != "c" {
		t.Fatalf("tabs=%#v", tabs)
	}
}

func TestQuickCheckDetectsInvalidDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.db")
	if err := os.WriteFile(path, []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Open(path)
	if err == nil {
		t.Fatal("expected open failure for invalid database")
	}
	if !IsCorruptError(err) {
		t.Fatalf("err=%v; want corrupt classification", err)
	}
}

func TestQuickCheckDetectsBitFlip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "flip.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AddHistory("deadbeef:/page/index.mu", "Page", "deadbeef", 1); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 128 {
		t.Fatal("database file too small for bit-flip test")
	}
	for i := 64; i < 128; i++ {
		raw[i] ^= 0xFF
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	database, err = Open(path)
	if err != nil {
		if IsCorruptError(err) {
			return
		}
		t.Fatalf("open err=%v", err)
	}
	defer database.Close()

	health, err := database.QuickCheck()
	if err != nil {
		if IsCorruptError(err) {
			return
		}
		t.Fatalf("quick_check err=%v", err)
	}
	if health.OK {
		t.Fatal("expected quick_check to report corruption after bit flip")
	}
}

func TestBackupAndRestore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.SaveTabs([]TabRow{{ID: "x", Title: "X", URL: "deadbeef:/page/x.mu"}}); err != nil {
		t.Fatal(err)
	}
	backupPath, err := database.BackupBesideDB()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	if err := RemoveFiles(path); err != nil {
		t.Fatal(err)
	}
	if err := RestoreBackup(path, backupPath); err != nil {
		t.Fatal(err)
	}

	database, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	tabs, err := database.Tabs()
	if err != nil {
		t.Fatal(err)
	}
	if len(tabs) != 1 || tabs[0].ID != "x" {
		t.Fatalf("tabs=%#v", tabs)
	}
}

func TestDeletePluginTransactional(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plugin.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err := database.UpsertPlugin("demo", true, `{}`); err != nil {
		t.Fatal(err)
	}
	if err := database.SetPluginSetting("demo", "k", "v"); err != nil {
		t.Fatal(err)
	}
	if err := database.DeletePlugin("demo"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.GetPlugin("demo"); !IsNotFound(err) {
		t.Fatalf("plugin row err=%v", err)
	}
	if _, err := database.GetPluginSetting("demo", "k"); !IsNotFound(err) {
		t.Fatalf("plugin setting err=%v", err)
	}
}

var errSimulatedWriteFailure = &simulatedWriteFailure{}

type simulatedWriteFailure struct{}

func (e *simulatedWriteFailure) Error() string { return "simulated write failure" }

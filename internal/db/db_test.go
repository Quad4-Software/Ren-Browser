// SPDX-License-Identifier: MIT
package db_test

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/db"
)

func TestSQLiteWALAndHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(path)
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

	if err := database.UpsertNode(db.NodeRow{
		Hash: "abc",
		Name: "Node",
	}); err != nil {
		t.Fatal(err)
	}
	nodes, err := database.ListNodes()
	if err != nil || len(nodes) != 1 {
		t.Fatalf("nodes=%#v err=%v", nodes, err)
	}

	if err := database.AddHistory("abc:/page/index.mu", "Node", "abc", 0); err != nil {
		t.Fatal(err)
	}
	hist, err := database.ListHistory(5)
	if err != nil || len(hist) != 1 {
		t.Fatalf("hist=%#v err=%v", hist, err)
	}
}

func TestStoredNodesAndHistoryAreCapped(t *testing.T) {
	t.Setenv("REN_BROWSER_MAX_STORED_NODES", "3")
	t.Setenv("REN_BROWSER_MAX_HISTORY_ROWS", "3")
	path := filepath.Join(t.TempDir(), "cap.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	for i := 1; i <= 8; i++ {
		if err := database.UpsertNode(db.NodeRow{
			Hash:     string(rune('a' + i - 1)),
			Name:     "n",
			LastSeen: int64(i),
		}); err != nil {
			t.Fatal(err)
		}
		if err := database.AddHistory("u", "t", "h", int64(i)); err != nil {
			t.Fatal(err)
		}
	}
	nodes, err := database.ListNodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 3 {
		t.Fatalf("nodes=%d want 3", len(nodes))
	}
	if nodes[0].LastSeen != 8 {
		t.Fatalf("newest last_seen=%d", nodes[0].LastSeen)
	}
	hist, err := database.ListHistory(20)
	if err != nil {
		t.Fatal(err)
	}
	if len(hist) != 3 {
		t.Fatalf("history=%d want 3", len(hist))
	}
}

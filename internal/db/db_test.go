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

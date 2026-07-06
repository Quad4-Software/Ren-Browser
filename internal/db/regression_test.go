// SPDX-License-Identifier: MIT
package db

import (
	"path/filepath"
	"testing"
)

func TestSQLiteBusyTimeoutConfigured(t *testing.T) {
	path := filepath.Join(t.TempDir(), "busy.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	var timeout int
	if err := database.sql.QueryRow(`PRAGMA busy_timeout`).Scan(&timeout); err != nil {
		t.Fatalf("pragma busy_timeout: %v", err)
	}
	if timeout != 5000 {
		t.Fatalf("busy_timeout=%d; want 5000", timeout)
	}
}

func TestSQLiteForeignKeysEnabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fk.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	var enabled int
	if err := database.sql.QueryRow(`PRAGMA foreign_keys`).Scan(&enabled); err != nil {
		t.Fatalf("pragma foreign_keys: %v", err)
	}
	if enabled != 1 {
		t.Fatalf("foreign_keys=%d; want 1", enabled)
	}
}

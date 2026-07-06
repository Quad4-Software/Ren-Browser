// SPDX-License-Identifier: MIT
package plugins_test

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/db"
	"renbrowser/internal/plugins"
)

func TestManagerCloseIdempotent(t *testing.T) {
	database, err := db.Open(filepath.Join(t.TempDir(), "plugins.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	manager := plugins.NewManager(database)
	if err := manager.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

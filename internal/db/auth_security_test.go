// SPDX-License-Identifier: MIT
package db_test

import (
	"path/filepath"
	"testing"
	"time"

	"renbrowser/internal/db"
)

func TestAuthSessionExpiryPrune(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err := database.CreateAuthSession("expired", time.Now().Add(-time.Hour).Unix()); err != nil {
		t.Fatal(err)
	}
	if err := database.CreateAuthSession("active", time.Now().Add(time.Hour).Unix()); err != nil {
		t.Fatal(err)
	}
	if err := database.PruneExpiredAuthSessions(); err != nil {
		t.Fatal(err)
	}
	if _, err := database.GetAuthSession("expired"); err == nil {
		t.Fatal("expected expired session to be pruned")
	}
	if _, err := database.GetAuthSession("active"); err != nil {
		t.Fatal("expected active session to remain")
	}
}

func TestAuthClearResetsState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clear.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err := database.SetAuthCredential("hash"); err != nil {
		t.Fatal(err)
	}
	if err := database.CreateAuthSession("tok", time.Now().Add(time.Hour).Unix()); err != nil {
		t.Fatal(err)
	}
	if err := database.UpsertAuthBruteState(db.AuthBruteState{ClientHash: "c", FailCount: 1}); err != nil {
		t.Fatal(err)
	}

	if err := database.ClearAuthCredential(); err != nil {
		t.Fatal(err)
	}
	if err := database.ClearAuthSessions(); err != nil {
		t.Fatal(err)
	}
	if err := database.ClearAuthBruteState(); err != nil {
		t.Fatal(err)
	}

	enabled, err := database.AuthEnabled()
	if err != nil || enabled {
		t.Fatalf("enabled=%v err=%v", enabled, err)
	}
	if _, err := database.GetAuthSession("tok"); err == nil {
		t.Fatal("expected sessions cleared")
	}
	if _, err := database.GetAuthBruteState("c"); err == nil {
		t.Fatal("expected brute state cleared")
	}
}

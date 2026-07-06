// SPDX-License-Identifier: MIT
package db_test

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/auth"
	"renbrowser/internal/db"
)

func TestAuthCredentialRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	enabled, err := database.AuthEnabled()
	if err != nil || enabled {
		t.Fatalf("enabled=%v err=%v", enabled, err)
	}

	hash, err := auth.HashPassword("test-pass")
	if err != nil {
		t.Fatal(err)
	}
	if err := database.SetAuthCredential(hash); err != nil {
		t.Fatal(err)
	}
	enabled, err = database.AuthEnabled()
	if err != nil || !enabled {
		t.Fatalf("enabled=%v err=%v", enabled, err)
	}

	if err := database.CreateAuthSession("abc", 9999999999); err != nil {
		t.Fatal(err)
	}
	if _, err := database.GetAuthSession("abc"); err != nil {
		t.Fatal(err)
	}
	if err := database.ClearAuthCredential(); err != nil {
		t.Fatal(err)
	}
}

func TestAuthBruteState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "brute.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	state := db.AuthBruteState{ClientHash: "deadbeef", FailCount: 2, BannedUntil: 42}
	if err := database.UpsertAuthBruteState(state); err != nil {
		t.Fatal(err)
	}
	got, err := database.GetAuthBruteState("deadbeef")
	if err != nil || got.FailCount != 2 || got.BannedUntil != 42 {
		t.Fatalf("state=%#v err=%v", got, err)
	}
	if err := database.MarkAuthClientTrusted("deadbeef"); err != nil {
		t.Fatal(err)
	}
	got, err = database.GetAuthBruteState("deadbeef")
	if err != nil || !got.Trusted || got.FailCount != 0 {
		t.Fatalf("trusted=%#v err=%v", got, err)
	}
}

// SPDX-License-Identifier: MIT
package store_test

import (
	"strings"
	"testing"

	"renbrowser/internal/store"
)

func TestStoreCloseIdempotent(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/profile.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

func TestSaveTabsPreservesMetadataWhenTruncating(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/profile.db")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	large := strings.Repeat("x", 300*1024)
	saved := st.SaveTabs([]store.TabSnapshot{{
		ID:          "tab-1",
		Title:       "Important",
		URL:         "deadbeef:/page/index.mu",
		HTML:        large,
		LastRaw:     large,
		ContentType: "micron",
		Error:       "connection_failed",
		DurationMs:  42,
	}})
	if saved[0].Title != "Important" {
		t.Fatalf("title=%q", saved[0].Title)
	}
	if saved[0].URL != "deadbeef:/page/index.mu" {
		t.Fatalf("url=%q", saved[0].URL)
	}
	if saved[0].Error != "connection_failed" {
		t.Fatalf("error=%q", saved[0].Error)
	}
	if saved[0].DurationMs != 42 {
		t.Fatalf("duration=%d", saved[0].DurationMs)
	}
}

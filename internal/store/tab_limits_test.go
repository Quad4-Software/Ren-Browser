// SPDX-License-Identifier: MIT
package store_test

import (
	"strings"
	"testing"

	"renbrowser/internal/store"
)

func TestSaveTabsTruncatesLargeFields(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/profile.db")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	large := strings.Repeat("x", 300*1024)
	saved := st.SaveTabs([]store.TabSnapshot{{
		ID:      "tab-1",
		Title:   "big",
		URL:     "nomad://abc/page/index.mu",
		HTML:    large,
		LastRaw: large,
	}})
	if len(saved[0].HTML) >= len(large) {
		t.Fatalf("html not truncated: len=%d", len(saved[0].HTML))
	}
	if len(saved[0].LastRaw) >= len(large) {
		t.Fatalf("lastRaw not truncated: len=%d", len(saved[0].LastRaw))
	}

	loaded := st.Tabs()
	if len(loaded) != 1 {
		t.Fatalf("tabs = %d", len(loaded))
	}
	if len(loaded[0].HTML) >= len(large) {
		t.Fatalf("persisted html not truncated")
	}
}

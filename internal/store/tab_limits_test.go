// SPDX-License-Identifier: MIT
package store_test

import (
	"strings"
	"testing"

	"renbrowser/internal/store"
)

func TestSaveTabsTruncatesOversizedBodies(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/profile.db")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	t.Setenv("REN_BROWSER_MAX_PAGE_BYTES", "1024")
	large := strings.Repeat("x", 4096)
	saved := st.SaveTabs([]store.TabSnapshot{{
		ID:      "tab-1",
		Title:   "big",
		URL:     "nomad://abc/page/index.mu",
		Active:  true,
		HTML:    large,
		LastRaw: large,
	}})
	if len(saved[0].HTML) != 1024 {
		t.Fatalf("html not truncated: len=%d", len(saved[0].HTML))
	}
	if len(saved[0].LastRaw) != 1024 {
		t.Fatalf("lastRaw not truncated: len=%d", len(saved[0].LastRaw))
	}

	loaded := st.Tabs()
	if len(loaded) != 1 {
		t.Fatalf("tabs = %d", len(loaded))
	}
	if len(loaded[0].HTML) != 1024 {
		t.Fatalf("persisted html not truncated: %d", len(loaded[0].HTML))
	}
}

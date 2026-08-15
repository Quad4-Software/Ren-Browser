// SPDX-License-Identifier: MIT
package store_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"renbrowser/internal/db"
	"renbrowser/internal/store"
)

func TestSaveTabsStoresBodiesOnDiskNotSQLite(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "profile.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	large := strings.Repeat("x", 300*1024)
	saved := st.SaveTabs([]store.TabSnapshot{
		{ID: "tab-1", Title: "big", URL: "nomad://abc/page/index.mu", Active: true, HTML: large, LastRaw: large},
		{ID: "tab-2", Title: "idle", URL: "nomad://abc/page/other.mu", HTML: large, LastRaw: large},
	})
	if len(saved[0].HTML) != len(large) {
		t.Fatalf("active html not kept for write: len=%d", len(saved[0].HTML))
	}

	loaded := st.Tabs()
	if len(loaded) != 2 {
		t.Fatalf("tabs = %d", len(loaded))
	}
	if loaded[0].HTML != large || loaded[0].LastRaw != large {
		t.Fatalf("active tab body missing: html=%d raw=%d", len(loaded[0].HTML), len(loaded[0].LastRaw))
	}
	if loaded[1].HTML != "" || loaded[1].LastRaw != "" {
		t.Fatalf("inactive tab should omit body: html=%d raw=%d", len(loaded[1].HTML), len(loaded[1].LastRaw))
	}

	all := st.TabsWithBodies()
	if all[1].HTML != large {
		t.Fatalf("TabsWithBodies missing idle html: %d", len(all[1].HTML))
	}

	raw, err := os.ReadFile(filepath.Join(dir, "profile.db"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), large[:64]) {
		t.Fatal("sqlite still contains page body")
	}
}

func TestSaveTabsKeepsExistingBodyWhenPayloadOmitsIt(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "profile.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	st.SaveTabs([]store.TabSnapshot{{
		ID: "tab-1", Title: "kept", URL: "nomad://abc/page/index.mu", Active: true, HTML: "<p>full</p>", LastRaw: "full",
	}})
	st.SaveTabs([]store.TabSnapshot{{
		ID: "tab-1", Title: "kept", URL: "nomad://abc/page/index.mu", Active: true,
	}})
	loaded := st.Tabs()
	if loaded[0].HTML != "<p>full</p>" || loaded[0].LastRaw != "full" {
		t.Fatalf("expected retained body, got html=%q raw=%q", loaded[0].HTML, loaded[0].LastRaw)
	}
}

func TestOpenMigratesLegacySQLiteBodiesToFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	html := "<article>legacy</article>"
	if err := database.SaveTabs([]db.TabRow{{
		ID: "legacy", Title: "Legacy", URL: "about:", Active: true, HTML: html, LastRaw: "src",
	}}); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	loaded := st.Tabs()
	if len(loaded) != 1 || loaded[0].HTML != html || loaded[0].LastRaw != "src" {
		t.Fatalf("migrated tabs=%#v", loaded)
	}

	rows, err := st.DB().Tabs()
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].HTML != "" || rows[0].LastRaw != "" {
		t.Fatalf("sqlite bodies not cleared: %#v", rows[0])
	}
}

func TestTabBodyFileStorageBoundsHeap(t *testing.T) {
	const (
		pageSize = 256 << 10
		tabs     = 24
	)
	st, err := store.Open(filepath.Join(t.TempDir(), "profile.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	body := strings.Repeat("m", pageSize)
	snapshots := make([]store.TabSnapshot, tabs)
	for i := range snapshots {
		snapshots[i] = store.TabSnapshot{
			ID:      "tab-" + string(rune('a'+i)),
			Title:   "t",
			URL:     "nomad://abc/page/index.mu",
			Active:  i == 0,
			HTML:    body,
			LastRaw: body,
		}
	}

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	st.SaveTabs(snapshots)
	loaded := st.Tabs()

	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	var loadedBytes int
	for _, tab := range loaded {
		loadedBytes += len(tab.HTML) + len(tab.LastRaw)
	}
	uncapped := tabs * pageSize * 2
	t.Logf("tabs=%d page=%dKiB uncapped=%dMiB loaded_bodies=%dKiB", tabs, pageSize>>10, uncapped>>20, loadedBytes>>10)
	t.Logf("heap_inuse before=%dMiB after=%dMiB delta=%dMiB",
		before.HeapInuse>>20, after.HeapInuse>>20, int64(after.HeapInuse-before.HeapInuse)>>20)

	if loadedBytes > pageSize*2+pageSize {
		t.Fatalf("GetTabs loaded too much body data: %d", loadedBytes)
	}
	if after.HeapInuse > before.HeapInuse+uint64(pageSize)*8+8<<20 {
		t.Fatalf("heap grew too much: before=%d after=%d", before.HeapInuse, after.HeapInuse)
	}
}

func TestEditorDraftReadsFileBody(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "profile.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	st.SaveTabs([]store.TabSnapshot{{
		ID: "ed", Title: "Editor", URL: "editor:", Active: false, LastRaw: ">>hello",
	}})
	if got := st.EditorDraft(); got != ">>hello" {
		t.Fatalf("draft=%q", got)
	}
}

func TestResetClearsTabBodyFiles(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "profile.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	st.SaveTabs([]store.TabSnapshot{{
		ID: "tab-1", Title: "A", URL: "about:", Active: true, HTML: "<p>x</p>",
	}})
	if err := st.Reset(); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "tab-pages"))
	if len(entries) != 0 {
		t.Fatalf("expected empty tab-pages after reset, got %d", len(entries))
	}
}

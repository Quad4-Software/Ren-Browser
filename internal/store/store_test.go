// SPDX-License-Identifier: MIT
package store_test

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/nomadnet"
	"renbrowser/internal/store"
)

func TestStoreFavoritesAndHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "renbrowser.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	url := "deadbeef:/page/index.mu"
	favs := s.AddFavorite(url)
	if len(favs) != 1 || favs[0] != url {
		t.Fatalf("favorites = %#v", favs)
	}

	if err := s.AddHistory(url, "Test Node", "deadbeef"); err != nil {
		t.Fatal(err)
	}
	recent := s.Recent()
	if len(recent) != 1 {
		t.Fatalf("recent = %#v", recent)
	}

	s2, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	if len(s2.Favorites()) != 1 {
		t.Fatal("favorites not persisted")
	}
	hist, err := s2.BrowsingHistory(10)
	if err != nil || len(hist) != 1 {
		t.Fatalf("history = %#v err=%v", hist, err)
	}
}

func TestStoreNodes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "renbrowser.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	node := nomadnet.Node{
		Hash:     "abb3ebcd03cb2388a838e70c001291f9",
		Name:     "Test Node",
		Hops:     2,
		LastSeen: 123,
	}
	if err := s.UpsertNode(node); err != nil {
		t.Fatal(err)
	}
	nodes, err := s.ListNodes()
	if err != nil || len(nodes) != 1 || nodes[0].Name != "Test Node" {
		t.Fatalf("nodes = %#v err=%v", nodes, err)
	}
}

func TestStorePinnedTabs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "renbrowser.db")
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	tabs := []store.TabSnapshot{
		{ID: "a", Title: "A", URL: "deadbeef:/page/a.mu", Active: true, Pinned: true},
		{ID: "b", Title: "B", URL: "deadbeef:/page/b.mu", Active: false},
	}
	s.SaveTabs(tabs)

	s2, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	loaded := s2.Tabs()
	if len(loaded) != 2 {
		t.Fatalf("tabs = %#v", loaded)
	}
	if !loaded[0].Pinned || loaded[0].ID != "a" {
		t.Fatalf("pinned tab = %#v", loaded[0])
	}
	if loaded[1].Pinned {
		t.Fatalf("unpinned tab = %#v", loaded[1])
	}
}

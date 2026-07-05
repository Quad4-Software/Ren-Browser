// SPDX-License-Identifier: MIT
package app

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/rns"
)

func TestBrowserPrefsUILanguagePersist(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "profile.db")

	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewBrowserServiceWithOptions(stack, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}

	want := svc.GetBrowserPrefs()
	want.UILanguage = "de"
	svc.SetBrowserPrefs(want)

	stack2, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewBrowserServiceWithOptions(stack2, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}

	got := reloaded.GetBrowserPrefs()
	if got.UILanguage != "de" {
		t.Fatalf("uiLanguage = %q, want de", got.UILanguage)
	}
}

func TestBrowserPrefsUILanguageDefaultsEmpty(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "profile.db")

	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewBrowserServiceWithOptions(stack, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}

	got := svc.GetBrowserPrefs()
	if got.UILanguage != "" {
		t.Fatalf("uiLanguage = %q, want empty default", got.UILanguage)
	}
}

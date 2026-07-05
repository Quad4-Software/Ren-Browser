// SPDX-License-Identifier: MIT
package app

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/rns"
)

func TestThemeDefaults(t *testing.T) {
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

	theme := svc.GetTheme()
	if theme.Accent != "#60a5fa" {
		t.Fatalf("accent = %q", theme.Accent)
	}
	if theme.Mode != "dark" {
		t.Fatalf("mode = %q", theme.Mode)
	}
}

func TestThemeSettingsPersist(t *testing.T) {
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

	want := ThemeSettings{
		Mode:           "light",
		Accent:         "#ff0000",
		FontFamily:     "Georgia, serif",
		FontSize:       16,
		CustomTokens:   map[string]string{"border": "#111111"},
		CompactToolbar: true,
	}
	svc.SetTheme(want)

	stack2, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewBrowserServiceWithOptions(stack2, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}

	got := reloaded.GetTheme()
	if got.Accent != want.Accent {
		t.Fatalf("accent = %q", got.Accent)
	}
	if got.Mode != want.Mode {
		t.Fatalf("mode = %q", got.Mode)
	}
	if got.FontFamily != want.FontFamily {
		t.Fatalf("fontFamily = %q", got.FontFamily)
	}
	if got.FontSize != want.FontSize {
		t.Fatalf("fontSize = %d", got.FontSize)
	}
	if got.CustomTokens["border"] != "#111111" {
		t.Fatalf("customTokens.border = %q", got.CustomTokens["border"])
	}
	if !got.CompactToolbar {
		t.Fatal("expected compactToolbar to persist")
	}
}

func TestResetSettingsTheme(t *testing.T) {
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

	svc.SetTheme(ThemeSettings{Mode: "light", Accent: "#ff0000", FontSize: 18})
	reset := svc.ResetSettings()
	if reset.Theme.Accent != "#60a5fa" {
		t.Fatalf("reset accent = %q", reset.Theme.Accent)
	}
	if reset.Theme.Mode != "dark" {
		t.Fatalf("reset mode = %q", reset.Theme.Mode)
	}

	stack2, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewBrowserServiceWithOptions(stack2, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.GetTheme()
	if got.Accent != "#60a5fa" {
		t.Fatalf("reloaded accent = %q", got.Accent)
	}
}

// SPDX-License-Identifier: MIT
package app_test

import (
	"runtime"
	"testing"

	"renbrowser/internal/app"
)

func TestBrowserPrefsUILanguagePersist(t *testing.T) {
	root := t.TempDir()
	svc := newTestServiceIn(t, root)

	want := svc.GetBrowserPrefs()
	want.UILanguage = "de"
	svc.SetBrowserPrefs(want)
	_ = svc.Store().Close()

	reloaded := newTestServiceIn(t, root)
	got := reloaded.GetBrowserPrefs()
	if got.UILanguage != "de" {
		t.Fatalf("uiLanguage = %q, want de", got.UILanguage)
	}
}

func TestBrowserPrefsUILanguageDefaultsEmpty(t *testing.T) {
	svc := newTestService(t)

	got := svc.GetBrowserPrefs()
	if got.UILanguage != "" {
		t.Fatalf("uiLanguage = %q, want empty default", got.UILanguage)
	}
}

func TestBrowserPrefsNativeTitlebarMissingUsesPlatformDefault(t *testing.T) {
	svc := newTestService(t)
	if err := svc.Store().SetSetting("browserPrefs", `{"openLinksInNewTab":true}`); err != nil {
		t.Fatal(err)
	}
	got := svc.GetBrowserPrefs()
	want := runtime.GOOS == "windows"
	if got.NativeTitlebar != want {
		t.Fatalf("nativeTitlebar = %v, want %v", got.NativeTitlebar, want)
	}
}

func TestDefaultBrowserPrefsNativeTitlebarMatchesPlatform(t *testing.T) {
	got := app.DefaultBrowserPrefs()
	want := runtime.GOOS == "windows"
	if got.NativeTitlebar != want {
		t.Fatalf("nativeTitlebar = %v, want %v", got.NativeTitlebar, want)
	}
}

func TestBrowserPrefsTabLayoutDefaultsTop(t *testing.T) {
	svc := newTestService(t)

	got := svc.GetBrowserPrefs()
	if got.TabLayout != "top" {
		t.Fatalf("tabLayout = %q, want top", got.TabLayout)
	}
}

func TestBrowserPrefsTabLayoutPersist(t *testing.T) {
	svc := newTestService(t)

	want := svc.GetBrowserPrefs()
	want.TabLayout = "left"
	svc.SetBrowserPrefs(want)
	if got := svc.GetBrowserPrefs(); got.TabLayout != "left" {
		t.Fatalf("tabLayout = %q, want left", got.TabLayout)
	}
}

func TestBrowserPrefsTabLayoutInvalidFallsBack(t *testing.T) {
	svc := newTestService(t)

	want := svc.GetBrowserPrefs()
	want.TabLayout = "sideways"
	svc.SetBrowserPrefs(want)
	if got := svc.GetBrowserPrefs(); got.TabLayout != "top" {
		t.Fatalf("tabLayout = %q, want top", got.TabLayout)
	}
}

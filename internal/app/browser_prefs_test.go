// SPDX-License-Identifier: MIT
package app_test

import (
	"testing"
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

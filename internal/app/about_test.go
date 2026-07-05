// SPDX-License-Identifier: MIT
package app_test

import (
	"testing"

	"renbrowser/internal/app"
)

func TestEditorURL(t *testing.T) {
	svc := newTestService(t)

	page := svc.Navigate("editor")
	if page.URL != "editor:" {
		t.Fatalf("url = %q", page.URL)
	}
	if page.ContentType != "editor" {
		t.Fatalf("content type = %q", page.ContentType)
	}
	if page.Raw == "" {
		t.Fatal("expected editor template")
	}
}

func TestConfigURL(t *testing.T) {
	svc := newTestService(t)

	page := svc.Navigate("config")
	if page.URL != "config:" {
		t.Fatalf("url = %q", page.URL)
	}
	if page.ContentType != "config" {
		t.Fatalf("content type = %q", page.ContentType)
	}
	if page.Raw == "" {
		t.Fatal("expected config text")
	}
}

func TestAboutURL(t *testing.T) {
	svc := newTestService(t)

	for _, raw := range []string{"about", "about:", "  About  "} {
		page := svc.Navigate(raw)
		if page.URL != "about:" {
			t.Fatalf("url = %q", page.URL)
		}
		if page.ContentType != "about" {
			t.Fatalf("content type = %q", page.ContentType)
		}
		if page.HTML == "" {
			t.Fatal("expected about html")
		}
	}
}

func TestLicenseURL(t *testing.T) {
	svc := newTestService(t)

	for _, raw := range []string{"license", "license:", "  License  "} {
		page := svc.Navigate(raw)
		if page.URL != "license:" {
			t.Fatalf("url = %q", page.URL)
		}
		if page.ContentType != "license" {
			t.Fatalf("content type = %q", page.ContentType)
		}
		if page.HTML == "" {
			t.Fatal("expected license html")
		}
	}
}

func TestKeybindDefaults(t *testing.T) {
	svc := newTestService(t)

	binds := svc.GetKeybinds()
	if binds.Bindings["focusUrl"] != "mod+l" {
		t.Fatalf("focusUrl = %q", binds.Bindings["focusUrl"])
	}

	updated := svc.SetKeybinds(app.KeybindSettings{
		Bindings: map[string]string{"reload": "mod+shift+r"},
	})
	if updated.Bindings["reload"] != "mod+shift+r" {
		t.Fatalf("reload = %q", updated.Bindings["reload"])
	}
	if updated.Bindings["focusUrl"] != "mod+l" {
		t.Fatalf("focusUrl not preserved: %q", updated.Bindings["focusUrl"])
	}
}

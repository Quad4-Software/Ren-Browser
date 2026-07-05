// SPDX-License-Identifier: MIT
package app_test

import (
	"strings"
	"testing"

	"renbrowser/internal/app"
	"renbrowser/internal/plugins"
)

func setupDocsService(t *testing.T) *app.BrowserService {
	t.Helper()
	svc := newTestService(t)
	svc.SetPluginManager(plugins.NewManager(svc.Store()))
	return svc
}

func TestDocsURL(t *testing.T) {
	svc := setupDocsService(t)

	picker := svc.Navigate("docs")
	if picker.ContentType != "docs" || picker.URL != "docs:" {
		t.Fatalf("picker url=%q type=%q", picker.URL, picker.ContentType)
	}

	for _, raw := range []string{"docs?lang=en", "docs:?lang=en&page=faq"} {
		page := svc.Navigate(raw)
		if page.ContentType != "docs" {
			t.Fatalf("raw=%q content type = %q", raw, page.ContentType)
		}
		if page.HTML == "" {
			t.Fatalf("raw=%q expected docs html", raw)
		}
	}

	prefs := svc.GetBrowserPrefs()
	if prefs.DocsLanguage != "en" {
		t.Fatalf("docsLanguage=%q", prefs.DocsLanguage)
	}

	page := svc.Navigate("docs")
	if page.URL != "docs:?lang=en" {
		t.Fatalf("remembered url=%q", page.URL)
	}
	if !strings.Contains(page.HTML, "Getting started") && !strings.Contains(page.HTML, "Table of contents") {
		t.Fatal("expected remembered english docs index")
	}
}

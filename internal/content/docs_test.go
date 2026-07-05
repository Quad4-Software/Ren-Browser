// SPDX-License-Identifier: MIT

package content_test

import (
	"strings"
	"testing"

	"renbrowser/internal/content"
)

func TestMatchDocsURL(t *testing.T) {
	for _, raw := range []string{"docs", "docs:", "docs?lang=en", "docs:?lang=en&page=faq", "  Docs  "} {
		if !content.MatchDocsURL(raw) {
			t.Fatalf("expected match for %q", raw)
		}
	}
	if content.MatchDocsURL("about:") {
		t.Fatal("about should not match docs")
	}
}

func TestParseDocsQuery(t *testing.T) {
	lang, page := content.ParseDocsQuery("docs:?lang=es&page=getting-started")
	if lang != "es" || page != "getting-started" {
		t.Fatalf("lang=%q page=%q", lang, page)
	}
	lang, page = content.ParseDocsQuery("docs?lang=EN")
	if lang != "en" || page != "" {
		t.Fatalf("lang=%q page=%q", lang, page)
	}
}

func TestSanitizeDocsPage(t *testing.T) {
	if got := content.SanitizeDocsPage("../secrets"); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := content.SanitizeDocsPage("faq.md"); got != "faq" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatDocsURL(t *testing.T) {
	if got := content.FormatDocsURL("en", "faq"); got != "docs:?lang=en&page=faq" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderDocsLanguagePicker(t *testing.T) {
	res, ok := content.RenderDocs(content.DocsRenderInput{RawURL: "docs:"})
	if !ok {
		t.Fatal("expected docs render")
	}
	if res.URL != "docs:" {
		t.Fatalf("url=%q", res.URL)
	}
	for _, lang := range content.SupportedDocsLangs() {
		if !strings.Contains(res.HTML, content.FormatDocsURL(lang, "")) {
			t.Fatalf("missing lang link %s", lang)
		}
	}
}

func TestRenderDocsUsesSavedLanguage(t *testing.T) {
	res, ok := content.RenderDocs(content.DocsRenderInput{
		RawURL:    "docs:",
		SavedLang: "en",
	})
	if !ok {
		t.Fatal("expected docs render")
	}
	if res.URL != "docs:?lang=en" {
		t.Fatalf("url=%q", res.URL)
	}
	if !strings.Contains(res.Raw, "Getting started") && !strings.Contains(res.Raw, "Table of contents") {
		t.Fatalf("expected english index markdown, got snippet=%q", res.Raw[:min(200, len(res.Raw))])
	}
}

func TestRenderDocsSavesLanguage(t *testing.T) {
	var saved string
	res, ok := content.RenderDocs(content.DocsRenderInput{
		RawURL: "docs:?lang=de",
		SaveLang: func(lang string) {
			saved = lang
		},
	})
	if !ok {
		t.Fatal("expected docs render")
	}
	if saved != "de" {
		t.Fatalf("saved=%q", saved)
	}
	if res.URL != "docs:?lang=de" {
		t.Fatalf("url=%q", res.URL)
	}
}

func TestRenderDocsPage(t *testing.T) {
	res, ok := content.RenderDocs(content.DocsRenderInput{
		RawURL:    "docs:?lang=en&page=faq",
		SavedLang: "ru",
		SaveLang: func(lang string) {
			if lang != "en" {
				t.Fatalf("lang=%q", lang)
			}
		},
	})
	if !ok {
		t.Fatal("expected docs render")
	}
	if res.Raw == "" {
		t.Fatal("expected markdown body")
	}
	if !strings.Contains(res.Raw, "# FAQ") && !strings.Contains(res.Raw, "FAQ") {
		t.Fatalf("missing FAQ heading in raw markdown")
	}
	if !strings.Contains(res.Raw, "installation.md") {
		t.Fatalf("expected relative markdown link in source")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

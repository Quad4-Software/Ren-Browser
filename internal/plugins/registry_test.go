// SPDX-License-Identifier: MIT
package plugins_test

import (
	"testing"

	"renbrowser/internal/nomadnet"
	"renbrowser/internal/plugins"
	"renbrowser/internal/plugins/builtin"
)

func TestRegistryRendererPriority(t *testing.T) {
	reg := plugins.NewRegistry()
	builtin.RegisterRenderers(reg)
	body := []byte("`>Title")
	kind := nomadnet.DetectContentType("/page/index.mu", body)
	r, ok := reg.BestRenderer("/page/index.mu", body, kind)
	if !ok {
		t.Fatal("expected renderer")
	}
	if r.ID() != "builtin.micron" {
		t.Fatalf("renderer = %s", r.ID())
	}
}

func TestRegistrySchemeAbout(t *testing.T) {
	reg := plugins.NewRegistry()
	builtin.RegisterSchemes(reg, builtin.SchemeDeps{
		AboutHTML: func() string { return "<p>about</p>" },
	})
	res, ok := reg.HandleScheme("about:")
	if !ok {
		t.Fatal("expected about scheme")
	}
	if res.ContentType != "about" {
		t.Fatalf("content type = %q", res.ContentType)
	}
}

func TestRegistrySchemeDocs(t *testing.T) {
	reg := plugins.NewRegistry()
	builtin.RegisterSchemes(reg, builtin.SchemeDeps{
		DocsPage: func(rawURL string) (plugins.SchemeResult, bool) {
			return plugins.SchemeResult{
				URL:          "docs:?lang=en",
				ContentType:  "docs",
				HTML:         "<p>docs</p>",
				HistoryTitle: "Documentation",
			}, true
		},
	})
	res, ok := reg.HandleScheme("docs:?lang=en")
	if !ok {
		t.Fatal("expected docs scheme")
	}
	if res.ContentType != "docs" {
		t.Fatalf("content type = %q", res.ContentType)
	}
}

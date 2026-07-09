// SPDX-License-Identifier: MIT
package regression_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"renbrowser/internal/apperrors"
	"renbrowser/internal/cache"
	"renbrowser/internal/content"
	"renbrowser/internal/limits"
	"renbrowser/internal/nomadnet"
	"renbrowser/internal/plugins/builtin"
	"renbrowser/internal/store"
)

func TestReliabilityPipelineRenderAndClassify(t *testing.T) {
	builtin.RegisterRenderers(content.RendererRegistry())

	body := []byte("`>Title\n[`link`:/page/other.mu`label]")
	rendered := content.Render("/page/index.mu", body, "abb3ebcd03cb2388a838e70c001291f9")
	if rendered.Kind != "micron" {
		t.Fatalf("kind=%q", rendered.Kind)
	}
	plain := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(rendered.HTML, "")
	if !strings.Contains(plain, "Title") {
		t.Fatal("expected rendered title")
	}
	if !strings.Contains(rendered.HTML, `class="Mu-mnt"`) {
		t.Fatal("expected force-monospace cells for ASCII alignment")
	}

	kind, _ := apperrors.ClassifyFetch("response too large: received 16 bytes (limit 8)", nil)
	if kind != apperrors.KindPayloadTooLarge {
		t.Fatalf("kind=%q", kind)
	}
}

func TestReliabilityPipelineCacheAndLimits(t *testing.T) {
	t.Setenv("REN_BROWSER_MAX_PAGE_BYTES", "1024")
	if limits.MaxPageBytes() != 1024 {
		t.Fatalf("max page=%d", limits.MaxPageBytes())
	}

	pageCache := cache.NewPageCache(4)
	req := nomadnet.RequestData{}
	pageCache.Put("node", "/page/index.mu", req, []byte("cached"), "micron")
	entry, ok := pageCache.Get("node", "/page/index.mu", req)
	if !ok || string(entry.Body) != "cached" {
		t.Fatalf("entry=%#v ok=%v", entry, ok)
	}

	err := nomadnet.CheckResponseSize(make([]byte, 16), 16, limits.MaxPageBytes())
	if err != nil {
		t.Fatalf("unexpected size error within limit: %v", err)
	}
	err = nomadnet.CheckResponseSize(make([]byte, 16), 16, 8)
	if err == nil {
		t.Fatal("expected payload limit error")
	}
}

func TestReliabilityPipelineStoreTruncation(t *testing.T) {
	st, err := store.Open(t.TempDir() + "/profile.db")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	large := strings.Repeat("z", 300*1024)
	saved := st.SaveTabs([]store.TabSnapshot{{
		ID:    "tab-1",
		Title: "Keep",
		URL:   "deadbeef:/page/index.mu",
		HTML:  large,
	}})
	if saved[0].Title != "Keep" {
		t.Fatalf("title=%q", saved[0].Title)
	}
	if len(saved[0].HTML) >= len(large) {
		t.Fatal("html should be truncated")
	}
}

func TestMain(m *testing.M) {
	_ = os.Setenv("REN_BROWSER_MAX_PAGE_BYTES", "")
	_ = os.Setenv("REN_BROWSER_MAX_FILE_BYTES", "")
	_ = os.Setenv("REN_BROWSER_MAX_ASSET_BYTES", "")
	_ = os.Setenv("REN_BROWSER_MAX_TAB_FIELD_BYTES", "")
	os.Exit(m.Run())
}

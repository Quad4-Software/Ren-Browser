// SPDX-License-Identifier: MIT
package nomadnet_test

import (
	"testing"

	"renbrowser/internal/limits"
	"renbrowser/internal/nomadnet"
)

func TestMaxFetchBytesDiffersForFiles(t *testing.T) {
	page := limits.MaxFetchBytes("/page/index.mu")
	file := limits.MaxFetchBytes("/file/music/song.mp3")
	if page <= 0 {
		t.Fatalf("page=%d; page limit should be positive", page)
	}
	if file != 0 {
		t.Fatalf("file=%d; file downloads should be unlimited by default", file)
	}
}

func TestParseURLPreservesCachedRequestSuffix(t *testing.T) {
	raw := "abb3ebcd03cb2388a838e70c001291f9:/page/form.mu`category=general|field.user=alice"
	parsed, err := nomadnet.ParseURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Request.CacheKeySuffix() == "" {
		t.Fatal("expected cached suffix")
	}
	again := parsed.Request.CacheKeySuffix()
	if again != parsed.Request.CacheKeySuffix() {
		t.Fatalf("suffix changed between calls: %q vs %q", again, parsed.Request.CacheKeySuffix())
	}
}

func TestDetectContentTypeMixedCasePaths(t *testing.T) {
	cases := map[string]string{
		"/page/index.MU":    "micron",
		"/page/readme.MD":   "markdown",
		"/page/site.HTML":   "html",
		"/page/notes.TXT":   "plaintext",
		"/file/archive.ZIP": "plaintext",
	}
	for path, want := range cases {
		got := nomadnet.DetectContentType(path, []byte("probe"))
		if got != want {
			t.Fatalf("%s: got %q want %q", path, got, want)
		}
	}
}

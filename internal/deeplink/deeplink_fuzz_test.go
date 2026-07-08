// SPDX-License-Identifier: MIT
package deeplink_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"renbrowser/internal/deeplink"
)

func FuzzUnwrap(f *testing.F) {
	seeds := []string{
		"",
		"about:",
		"renbrowser:about",
		"renbrowser://open?url=about%3A",
		"renbrowser://abb3ebcd03cb2388a838e70c001291f9/page/index.mu",
		"rns://abb3ebcd03cb2388a838e70c001291f9/page/home.mu",
		"abb3ebcd03cb2388a838e70c001291f9",
		"https://example.com",
		"javascript:alert(1)",
		"renbrowser://open?url=https%3A%2F%2Fevil.test",
		"document:/x.pdf",
		"docs:?page=faq",
		"renbrowser://rns/abb3ebcd03cb2388a838e70c001291f9/page/x.mu",
		strings.Repeat("a", 4096),
		"renbrowser:\x00about",
		"renbrowser://open?url=" + strings.Repeat("%41", 200),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		target, ok := deeplink.Unwrap(raw)
		if !ok {
			return
		}
		if target == "" {
			t.Fatalf("ok with empty target for %q", raw)
		}
		if strings.IndexByte(target, 0) >= 0 {
			t.Fatalf("NUL in target %q", target)
		}
		if !utf8.ValidString(target) {
			t.Fatalf("invalid utf8 target from %q", raw)
		}
		lower := strings.ToLower(target)
		for _, blocked := range []string{"http:", "https:", "javascript:", "data:", "file:", "blob:", "ftp:", "mailto:"} {
			if strings.HasPrefix(lower, blocked) {
				t.Fatalf("unwrapped blocked scheme %q from %q", target, raw)
			}
		}
		// Idempotent for already-internal targets.
		again, ok2 := deeplink.Unwrap(target)
		if !ok2 || again == "" {
			t.Fatalf("re-unwrap failed for %q from %q", target, raw)
		}
	})
}

func FuzzHandleIncoming(f *testing.F) {
	f.Add("renbrowser:about")
	f.Add("https://example.com")
	f.Add("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu")
	f.Fuzz(func(t *testing.T, raw string) {
		deeplink.ClearPending()
		deeplink.SetHandler(nil)
		target, ok := deeplink.HandleIncoming(raw)
		if !ok {
			if deeplink.PeekPending() != "" {
				t.Fatal("pending set on rejection")
			}
			return
		}
		if target == "" {
			t.Fatal("empty accepted target")
		}
		if took := deeplink.TakePending(); took != target {
			t.Fatalf("pending %q != target %q", took, target)
		}
	})
}

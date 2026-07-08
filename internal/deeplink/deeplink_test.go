// SPDX-License-Identifier: MIT
package deeplink_test

import (
	"strings"
	"testing"

	"renbrowser/internal/deeplink"
)

func TestUnwrapRenBrowserForms(t *testing.T) {
	hash := "abb3ebcd03cb2388a838e70c001291f9"
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"", "", false},
		{"   ", "", false},
		{"renbrowser:", "", false},
		{"renbrowser://", "", false},
		{"renbrowser:about", "about:", true},
		{"renbrowser:about:", "about:", true},
		{"RENBROWSER:LICENSE", "license:", true},
		{"renbrowser://about", "about:", true},
		{"renbrowser://settings", "settings:", true},
		{"renbrowser://docs?page=faq", "docs:?page=faq", true},
		{"renbrowser:docs:?lang=en&page=faq", "docs:?lang=en&page=faq", true},
		{"renbrowser://open?url=about%3A", "about:", true},
		{"renbrowser://open?url=" + hash + "%3A%2Fpage%2Findex.mu", hash + ":/page/index.mu", true},
		{"renbrowser:open?url=license%3A", "license:", true},
		{"renbrowser:url=editor%3A", "editor:", true},
		{"renbrowser://" + hash, hash + ":/page/index.mu", true},
		{"renbrowser://" + hash + "/page/about.mu", hash + ":/page/about.mu", true},
		{"renbrowser://" + strings.ToUpper(hash) + "/page/x.mu?a=1", hash + ":/page/x.mu?a=1", true},
		{"renbrowser://rns/" + hash + "/page/home.mu", "rns://" + hash + "/page/home.mu", true},
		{"renbrowser://open?url=rns%3A%2F%2F" + hash + "%2Fpage%2Fhome.mu", "rns://" + hash + "/page/home.mu", true},
		{"renbrowser://open?url=renbrowser%3Aabout", "about:", true},
		{"rns://" + hash + "/page/index.mu", "rns://" + hash + "/page/index.mu", true},
		{hash, hash + ":/page/index.mu", true},
		{hash + ":/page/form.mu`user=alice", hash + ":/page/form.mu`user=alice", true},
		{"about:", "about:", true},
		{"ABOUT", "about:", true},
		{"docs?page=x", "docs:?page=x", true},
		{"document:book.epub", "document:/book.epub", true},
		{"document:/book.epub", "document:/book.epub", true},
		{"https://example.com", "", false},
		{"http://evil.test", "", false},
		{"javascript:alert(1)", "", false},
		{"data:text/html,hi", "", false},
		{"file:///etc/passwd", "", false},
		{"renbrowser://open?url=https%3A%2F%2Fevil.test", "", false},
		{"renbrowser://open?url=", "", false},
		{"ftp://example.com", "", false},
		{"not-a-url", "", false},
		{"aabb", "", false},
	}
	for _, tc := range cases {
		got, ok := deeplink.Unwrap(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("Unwrap(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestUnwrapRejectsNUL(t *testing.T) {
	if _, ok := deeplink.Unwrap("about:\x00evil"); ok {
		t.Fatal("expected NUL rejection")
	}
}

func TestHandleIncomingQueuesAndNotifies(t *testing.T) {
	deeplink.ClearPending()
	defer deeplink.ClearPending()
	deeplink.SetHandler(nil)

	var seen string
	deeplink.SetHandler(func(target string) { seen = target })
	defer deeplink.SetHandler(nil)

	target, ok := deeplink.HandleIncoming("renbrowser:about")
	if !ok || target != "about:" {
		t.Fatalf("HandleIncoming = (%q, %v)", target, ok)
	}
	if seen != "about:" {
		t.Fatalf("handler saw %q", seen)
	}
	if peek := deeplink.PeekPending(); peek != "about:" {
		t.Fatalf("PeekPending = %q", peek)
	}
	if took := deeplink.TakePending(); took != "about:" {
		t.Fatalf("TakePending = %q", took)
	}
	if again := deeplink.TakePending(); again != "" {
		t.Fatalf("second TakePending = %q", again)
	}
}

func TestHandleIncomingRejectsExternal(t *testing.T) {
	deeplink.ClearPending()
	defer deeplink.ClearPending()
	if _, ok := deeplink.HandleIncoming("https://example.com"); ok {
		t.Fatal("expected rejection")
	}
	if deeplink.PeekPending() != "" {
		t.Fatal("pending should stay empty")
	}
}

func TestEnqueueOverwrites(t *testing.T) {
	deeplink.ClearPending()
	defer deeplink.ClearPending()
	deeplink.Enqueue("about:")
	deeplink.Enqueue("license:")
	if got := deeplink.TakePending(); got != "license:" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractFromArgs(t *testing.T) {
	target, ok := deeplink.ExtractFromArgs([]string{"renbrowser", "--profile", "dev", "renbrowser:about"})
	if !ok || target != "about:" {
		t.Fatalf("ExtractFromArgs = (%q, %v)", target, ok)
	}
	if _, ok := deeplink.ExtractFromArgs([]string{"renbrowser", "-config", "/tmp/x"}); ok {
		t.Fatal("expected no deeplink")
	}
	if _, ok := deeplink.ExtractFromArgs([]string{"renbrowser", "https://example.com"}); ok {
		t.Fatal("expected external rejection")
	}
}

func TestEdgeCaseWhitespaceAndCase(t *testing.T) {
	got, ok := deeplink.Unwrap("  RenBrowser://About  ")
	if !ok || got != "about:" {
		t.Fatalf("got (%q, %v)", got, ok)
	}
	hash := "ABB3EBCD03CB2388A838E70C001291F9"
	got, ok = deeplink.Unwrap("  " + hash + "  ")
	if !ok || got != strings.ToLower(hash)+":/page/index.mu" {
		t.Fatalf("got (%q, %v)", got, ok)
	}
}

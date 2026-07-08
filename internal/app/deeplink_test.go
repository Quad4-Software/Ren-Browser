// SPDX-License-Identifier: MIT
package app_test

import (
	"testing"

	"renbrowser/internal/app"
	"renbrowser/internal/deeplink"
)

func TestHandleDeepLinkAcceptsAndQueues(t *testing.T) {
	deeplink.ClearPending()
	defer deeplink.ClearPending()

	svc := &app.BrowserService{}
	got := svc.HandleDeepLink("renbrowser:about")
	if got != "about:" {
		t.Fatalf("HandleDeepLink = %q", got)
	}
	if pending := svc.TakePendingDeepLink(); pending != "about:" {
		t.Fatalf("TakePendingDeepLink = %q", pending)
	}
	if again := svc.TakePendingDeepLink(); again != "" {
		t.Fatalf("second take = %q", again)
	}
}

func TestHandleDeepLinkRejectsExternal(t *testing.T) {
	deeplink.ClearPending()
	defer deeplink.ClearPending()

	svc := &app.BrowserService{}
	if got := svc.HandleDeepLink("https://example.com"); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	if pending := svc.TakePendingDeepLink(); pending != "" {
		t.Fatalf("pending = %q", pending)
	}
}

func TestHandleDeepLinkMeshAndRNS(t *testing.T) {
	deeplink.ClearPending()
	defer deeplink.ClearPending()

	hash := "abb3ebcd03cb2388a838e70c001291f9"
	svc := &app.BrowserService{}
	cases := []struct {
		in   string
		want string
	}{
		{"renbrowser://" + hash + "/page/x.mu", hash + ":/page/x.mu"},
		{"rns://" + hash + "/page/home.mu", "rns://" + hash + "/page/home.mu"},
		{"renbrowser://open?url=" + hash, hash + ":/page/index.mu"},
	}
	for _, tc := range cases {
		deeplink.ClearPending()
		got := svc.HandleDeepLink(tc.in)
		if got != tc.want {
			t.Fatalf("HandleDeepLink(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

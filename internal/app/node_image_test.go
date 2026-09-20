// SPDX-License-Identifier: MIT
package app_test

import (
	"strings"
	"testing"
)

func TestFetchNodeImageRejectsBadInput(t *testing.T) {
	svc := newTestServiceIn(t, t.TempDir())
	node := "abcdef0123456789abcdef0123456789"

	// Rejected paths fail before any network activity.
	for _, tc := range []string{
		"/page/index.mu",
		"/media/x.svg",
		"/file/x.png",
		"/file/x.jpg",
		"/media/../secret.png",
		"/media/x.exe`img=1",
		"/media/x.exe",
	} {
		_, err := svc.FetchNodeImage(node+":"+tc, "", "", false)
		if err == nil || strings.Contains(err.Error(), "reticulum not ready") {
			t.Fatalf("%q: expected path rejection, got %v", tc, err)
		}
	}

	if _, err := svc.FetchNodeImage("nothex:/media/x.png", "", "", false); err == nil {
		t.Fatal("expected invalid node hash rejection")
	}
}

func TestBrowserPrefsMicronImagesDefaults(t *testing.T) {
	svc := newTestServiceIn(t, t.TempDir())
	prefs := svc.GetBrowserPrefs()
	if prefs.MicronImagesMode != "ask" {
		t.Fatalf("default mode=%q want ask", prefs.MicronImagesMode)
	}
	if len(prefs.MicronImageNodes) != 0 {
		t.Fatalf("default node policies should be empty, got %v", prefs.MicronImageNodes)
	}
}

func TestBrowserPrefsMicronImagesMerge(t *testing.T) {
	svc := newTestServiceIn(t, t.TempDir())
	prefs := svc.GetBrowserPrefs()
	prefs.MicronImagesMode = "always"
	prefs.MicronImageNodes = map[string]string{
		"ABCDEF0123456789ABCDEF0123456789": "always",
		"11111111111111111111111111111111": "never",
		"not-a-hash":                       "always",
		"22222222222222222222222222222222": "bogus",
	}
	merged := svc.SetBrowserPrefs(prefs)
	if merged.MicronImagesMode != "always" {
		t.Fatalf("mode=%q", merged.MicronImagesMode)
	}
	if got := merged.MicronImageNodes["abcdef0123456789abcdef0123456789"]; got != "always" {
		t.Fatalf("expected normalized allow policy, got %q", got)
	}
	if got := merged.MicronImageNodes["11111111111111111111111111111111"]; got != "never" {
		t.Fatalf("expected never policy, got %q", got)
	}
	if _, ok := merged.MicronImageNodes["not-a-hash"]; ok {
		t.Fatal("invalid node hash must be dropped")
	}
	if _, ok := merged.MicronImageNodes["22222222222222222222222222222222"]; ok {
		t.Fatal("invalid policy value must be dropped")
	}

	prefs = svc.GetBrowserPrefs()
	prefs.MicronImagesMode = "nonsense"
	merged = svc.SetBrowserPrefs(prefs)
	if merged.MicronImagesMode != "ask" {
		t.Fatalf("invalid mode should reset to ask, got %q", merged.MicronImagesMode)
	}
}

func TestBrowserPrefsMicronImageNodesClear(t *testing.T) {
	svc := newTestServiceIn(t, t.TempDir())
	prefs := svc.GetBrowserPrefs()
	prefs.MicronImageNodes = map[string]string{"abcdef0123456789abcdef0123456789": "always"}
	svc.SetBrowserPrefs(prefs)
	prefs = svc.GetBrowserPrefs()
	prefs.MicronImageNodes = nil
	merged := svc.SetBrowserPrefs(prefs)
	if len(merged.MicronImageNodes) != 0 {
		t.Fatalf("clearing node policies failed: %v", merged.MicronImageNodes)
	}
}

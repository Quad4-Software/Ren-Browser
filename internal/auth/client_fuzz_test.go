// SPDX-License-Identifier: MIT
package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"renbrowser/internal/auth"
)

func FuzzNormalizeIP(f *testing.F) {
	f.Add("127.0.0.1")
	f.Add("[::1]:8080")
	f.Add("203.0.113.1:443")
	f.Add("")
	f.Add("not-an-ip")

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 512 {
			t.Skip("skip oversized input")
		}
		out := auth.NormalizeIP(raw)
		if len(out) > len(raw)+16 {
			t.Fatalf("normalized longer than input: %q -> %q", raw, out)
		}
		_ = auth.ClientHash("pepper", out, "ua")
	})
}

func FuzzClientHash(f *testing.F) {
	f.Add("pepper", "127.0.0.1", "Mozilla/5.0")
	f.Add("", "::1", "")
	f.Add("pepper", "[2001:db8::1]:443", "bot")

	f.Fuzz(func(t *testing.T, pepper, ip, ua string) {
		if len(pepper) > 256 || len(ip) > 256 || len(ua) > 2048 {
			t.Skip("skip oversized input")
		}
		h := auth.ClientHash(pepper, ip, ua)
		if len(h) != 64 {
			t.Fatalf("hash length = %d", len(h))
		}
		for _, c := range h {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				t.Fatalf("non-hex in hash: %q", h)
			}
		}
	})
}

func FuzzClientIP(f *testing.F) {
	f.Add(false, "127.0.0.1:8080", "", "")
	f.Add(true, "10.0.0.1:8080", "203.0.113.5, 198.51.100.2", "")
	f.Add(true, "10.0.0.1:8080", "", "198.51.100.9")

	f.Fuzz(func(t *testing.T, trust bool, remote, forwarded, realIP string) {
		if len(remote) > 256 || len(forwarded) > 1024 || len(realIP) > 256 {
			t.Skip("skip oversized input")
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = remote
		if forwarded != "" {
			req.Header.Set("X-Forwarded-For", forwarded)
		}
		if realIP != "" {
			req.Header.Set("X-Real-IP", realIP)
		}
		_ = auth.ClientIP(req, trust)
	})
}

func TestClientHashDiffersByUserAgent(t *testing.T) {
	a := auth.ClientHash("pepper", "1.2.3.4", "agent-a")
	b := auth.ClientHash("pepper", "1.2.3.4", "agent-b")
	if a == b {
		t.Fatal("expected different hashes for different user agents")
	}
}

func TestClientIPTrustProxy(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 198.51.100.1")

	if got := auth.ClientIP(req, false); got != "10.0.0.1" {
		t.Fatalf("without trust = %q", got)
	}
	if got := auth.ClientIP(req, true); got != "203.0.113.7" {
		t.Fatalf("with trust = %q", got)
	}
}

func TestNormalizeIPv6Compressed(t *testing.T) {
	cases := []struct{ in, want string }{
		{"::1", "::1"},
		{"[::1]:8080", "::1"},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:db8:85a3::8a2e:370:7334"},
	}
	for _, tc := range cases {
		if got := auth.NormalizeIP(tc.in); got != tc.want {
			t.Fatalf("NormalizeIP(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestClientHashDoesNotContainRawIP(t *testing.T) {
	ip := "203.0.113.44"
	h := auth.ClientHash("pepper", ip, "Mozilla")
	if strings.Contains(h, ip) {
		t.Fatal("hash must not contain raw IP")
	}
}

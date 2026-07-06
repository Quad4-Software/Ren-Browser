// SPDX-License-Identifier: MIT
package auth_test

import (
	"testing"

	"renbrowser/internal/auth"
)

func FuzzParseWhitelist(f *testing.F) {
	f.Add("127.0.0.1,::1")
	f.Add("192.168.0.0/16,2001:db8::/32")
	f.Add("garbage,not-ip")
	f.Add("")
	f.Add("10.0.0.1/8,10.0.0.1/8")

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 4096 {
			t.Skip("skip oversized input")
		}
		w, err := auth.ParseWhitelist(raw)
		if err != nil {
			return
		}
		probeIPs := []string{
			"127.0.0.1",
			"::1",
			"10.0.0.1",
			"203.0.113.1",
			"2001:db8::1",
			"not-an-ip",
			"",
		}
		for _, ip := range probeIPs {
			_ = w.Allows(ip)
		}
	})
}

func FuzzWhitelistAllows(f *testing.F) {
	f.Add("127.0.0.1", "127.0.0.1")
	f.Add("192.168.0.0/16", "192.168.55.10")
	f.Add("::1", "::1")

	f.Fuzz(func(t *testing.T, list, probe string) {
		if len(list) > 1024 || len(probe) > 256 {
			t.Skip("skip oversized input")
		}
		w, err := auth.ParseWhitelist(list)
		if err != nil {
			return
		}
		allowed := w.Allows(probe)
		if allowed {
			allowed2 := w.Allows(probe)
			if !allowed2 {
				t.Fatal("Allows must be deterministic")
			}
		}
	})
}

func TestWhitelistNilDenies(t *testing.T) {
	var w *auth.Whitelist
	if w.Allows("127.0.0.1") {
		t.Fatal("nil whitelist should deny")
	}
}

func TestWhitelistIPv6CIDR(t *testing.T) {
	w, err := auth.ParseWhitelist("2001:db8::/32")
	if err != nil {
		t.Fatal(err)
	}
	if !w.Allows("2001:db8::1") {
		t.Fatal("expected match inside IPv6 CIDR")
	}
	if w.Allows("2001:db9::1") {
		t.Fatal("expected outside CIDR to be denied")
	}
}

func TestParseWhitelistRejectsInvalid(t *testing.T) {
	if _, err := auth.ParseWhitelist("not-an-ip"); err == nil {
		t.Fatal("expected error")
	}
	if _, err := auth.ParseWhitelist("10.0.0.0/99"); err == nil {
		t.Fatal("expected error for bad CIDR")
	}
}

func TestParseWhitelistIgnoresEmptyEntries(t *testing.T) {
	w, err := auth.ParseWhitelist("127.0.0.1,, ,::1")
	if err != nil {
		t.Fatal(err)
	}
	if !w.Allows("127.0.0.1") || !w.Allows("::1") {
		t.Fatal("expected valid entries to parse")
	}
}

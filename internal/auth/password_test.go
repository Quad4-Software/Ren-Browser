// SPDX-License-Identifier: MIT
package auth_test

import (
	"strings"
	"testing"

	"renbrowser/internal/auth"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := auth.HashPassword("mesh-secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.VerifyPassword(hash, "mesh-secret"); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := auth.VerifyPassword(hash, "wrong"); err == nil {
		t.Fatal("expected invalid password")
	}
}

func TestHashPasswordUniqueSalts(t *testing.T) {
	a, err := auth.HashPassword("same")
	if err != nil {
		t.Fatal(err)
	}
	b, err := auth.HashPassword("same")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("expected unique salts for identical passwords")
	}
	if err := auth.VerifyPassword(a, "same"); err != nil {
		t.Fatal(err)
	}
	if err := auth.VerifyPassword(b, "same"); err != nil {
		t.Fatal(err)
	}
}

func TestHashPasswordRejectsEmpty(t *testing.T) {
	if _, err := auth.HashPassword(""); err == nil {
		t.Fatal("expected error")
	}
	if _, err := auth.HashPassword("   "); err == nil {
		t.Fatal("expected error for whitespace")
	}
}

func TestVerifyPasswordRejectsTamperedHash(t *testing.T) {
	hash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(hash, "$")
	parts[len(parts)-1] = parts[len(parts)-1] + "ff"
	tampered := strings.Join(parts, "$")
	if err := auth.VerifyPassword(tampered, "secret"); err == nil {
		t.Fatal("expected tampered hash to fail")
	}
}

func TestVerifyPasswordRejectsGarbage(t *testing.T) {
	cases := []string{
		"",
		"bcrypt-like",
		"$argon2id$v=99$m=1,t=1,p=1$abc$def",
		"$argon2id$v=19$m=65536,t=3,p=2$!!!$!!!",
	}
	for _, c := range cases {
		if err := auth.VerifyPassword(c, "pw"); err == nil {
			t.Fatalf("expected rejection for %q", c)
		}
	}
}

func TestNewSessionTokenUnique(t *testing.T) {
	seen := make(map[string]struct{}, 32)
	for i := 0; i < 32; i++ {
		token, err := auth.NewSessionToken()
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := seen[token]; ok {
			t.Fatal("duplicate session token")
		}
		seen[token] = struct{}{}
	}
}

func TestParsePositiveInt(t *testing.T) {
	if got := auth.ParsePositiveInt("5", 1); got != 5 {
		t.Fatalf("got %d", got)
	}
	if got := auth.ParsePositiveInt("bad", 3); got != 3 {
		t.Fatalf("got %d", got)
	}
	if got := auth.ParsePositiveInt("-1", 3); got != 3 {
		t.Fatalf("got %d", got)
	}
}

func TestClientHashStableForIPv6(t *testing.T) {
	a := auth.ClientHash("pepper", "::1", "Mozilla/5.0")
	b := auth.ClientHash("pepper", "[::1]:1234", "Mozilla/5.0")
	if a != b {
		t.Fatalf("hashes differ: %q vs %q", a, b)
	}
}

func TestWhitelistIPv4AndCIDR(t *testing.T) {
	w, err := auth.ParseWhitelist("127.0.0.1,192.168.0.0/16,::1")
	if err != nil {
		t.Fatal(err)
	}
	if !w.Allows("127.0.0.1") || !w.Allows("192.168.1.42") || !w.Allows("::1") {
		t.Fatal("expected whitelist match")
	}
	if w.Allows("10.0.0.1") {
		t.Fatal("unexpected whitelist match")
	}
}

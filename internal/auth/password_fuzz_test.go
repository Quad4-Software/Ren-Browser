// SPDX-License-Identifier: MIT
package auth_test

import (
	"strings"
	"testing"

	"renbrowser/internal/auth"
)

func FuzzVerifyPassword(f *testing.F) {
	hash, err := auth.HashPassword("seed-password")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(hash, "seed-password")
	f.Add("not-a-valid-hash", "anything")
	f.Add("$argon2id$v=19$m=65536,t=3,p=2$c2FsdHNhbHRzYWx0$hashhashhash", "pw")
	f.Add("", "")
	f.Add(strings.Repeat("A", 4096), "pw")

	f.Fuzz(func(t *testing.T, encoded, password string) {
		if len(encoded) > 8192 || len(password) > 8192 {
			t.Skip("skip oversized input")
		}
		err := auth.VerifyPassword(encoded, password)
		if encoded == hash && password == "seed-password" && err != nil {
			t.Fatalf("valid hash rejected: %v", err)
		}
	})
}

func FuzzHashPassword(f *testing.F) {
	f.Add("normal")
	f.Add("")
	f.Add("   ")
	f.Add(strings.Repeat("x", 1024))

	f.Fuzz(func(t *testing.T, password string) {
		if len(password) > 4096 {
			t.Skip("skip oversized password")
		}
		hash, err := auth.HashPassword(password)
		if strings.TrimSpace(password) == "" {
			if err == nil {
				t.Fatal("expected error for empty password")
			}
			return
		}
		if err != nil {
			t.Fatalf("hash: %v", err)
		}
		if !strings.HasPrefix(hash, "$argon2id$") {
			t.Fatalf("unexpected format: %q", hash)
		}
		if err := auth.VerifyPassword(hash, password); err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if err := auth.VerifyPassword(hash, password+" "); err == nil {
			t.Fatal("accepted modified password")
		}
	})
}

func FuzzNewSessionToken(f *testing.F) {
	f.Add(byte(0))

	f.Fuzz(func(t *testing.T, _ byte) {
		token, err := auth.NewSessionToken()
		if err != nil {
			t.Fatal(err)
		}
		if len(token) < 32 {
			t.Fatalf("token too short: %d", len(token))
		}
		if strings.ContainsAny(token, "+/=") {
			t.Fatalf("token should be URL-safe: %q", token)
		}
	})
}

// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeDirIntegrityHashStable(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("renbrowser.plugin.json", `{"manifestVersion":1,"id":"test","name":"Test","version":"1.0.0"}`)
	write("main.js", "console.log('ok');")

	first, err := ComputeDirIntegrityHash(dir)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	second, err := ComputeDirIntegrityHash(dir)
	if err != nil {
		t.Fatalf("hash again: %v", err)
	}
	if first != second || first == "" {
		t.Fatalf("expected stable non-empty hash, got %q and %q", first, second)
	}

	write("main.js", "console.log('changed');")
	third, err := ComputeDirIntegrityHash(dir)
	if err != nil {
		t.Fatalf("hash changed file: %v", err)
	}
	if third == first {
		t.Fatal("expected hash to change after file modification")
	}
}

func TestVerifyDirIntegrityDetectsTamper(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "renbrowser.plugin.json"), []byte(`{"manifestVersion":1,"id":"test","name":"Test","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := ComputeDirIntegrityHash(dir)
	if err != nil {
		t.Fatal(err)
	}
	ok, _, err := VerifyDirIntegrity(dir, hash)
	if err != nil || !ok {
		t.Fatalf("verify: ok=%v err=%v", ok, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "extra.js"), []byte("evil"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, _, err = VerifyDirIntegrity(dir, hash)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected tamper detection")
	}
}

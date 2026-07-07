// SPDX-License-Identifier: MIT
package plugins_test

import (
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/plugins"
)

func TestDirSizeBytes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "nested")
	if err := os.MkdirAll(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.bin"), make([]byte, 1024), 0o600); err != nil {
		t.Fatal(err)
	}

	size, err := plugins.DirSizeBytes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if size != int64(5+1024) {
		t.Fatalf("size=%d", size)
	}
}

func TestDirSizeBytesMissingDir(t *testing.T) {
	_, err := plugins.DirSizeBytes(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

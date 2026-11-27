// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCopyDirRejectsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires extra privileges on Windows")
	}
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "dest")
	if err := os.WriteFile(filepath.Join(src, "renbrowser.plugin.json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(secret, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, filepath.Join(src, "linked")); err != nil {
		t.Fatal(err)
	}
	if err := copyDir(src, dest); err == nil {
		t.Fatal("expected symlink copy to be rejected")
	}
}

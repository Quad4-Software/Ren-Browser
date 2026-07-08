// SPDX-License-Identifier: MIT
package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDataRootDefault(t *testing.T) {
	// Reset dataRoot
	oldDataRoot := dataRoot
	dataRoot = ""
	defer func() {
		dataRoot = oldDataRoot
	}()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("skipping test; UserHomeDir not available")
	}

	got := DataRoot()
	if got != home {
		t.Errorf("DataRoot() = %q; want %q", got, home)
	}
}

func TestSetDataRoot(t *testing.T) {
	oldDataRoot := dataRoot
	dataRoot = ""
	defer func() {
		dataRoot = oldDataRoot
	}()

	want := "/tmp/custom-data-root"
	SetDataRoot(want)

	got := DataRoot()
	if got != want {
		t.Errorf("DataRoot() after SetDataRoot = %q; want %q", got, want)
	}
}

func TestJoin(t *testing.T) {
	oldDataRoot := dataRoot
	dataRoot = ""
	defer func() {
		dataRoot = oldDataRoot
	}()

	SetDataRoot("/tmp/custom-data-root")
	got := Join("subdir", "file.txt")
	want := filepath.Clean("/tmp/custom-data-root/subdir/file.txt")

	if got != want {
		t.Errorf("Join() = %q; want %q", got, want)
	}
}

func TestUserDownloadDirEmptyOnDesktop(t *testing.T) {
	if got := UserDownloadDir(); got != "" {
		t.Fatalf("UserDownloadDir() = %q; want empty on non-android/non-ios", got)
	}
}

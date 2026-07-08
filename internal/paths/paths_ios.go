//go:build ios

// SPDX-License-Identifier: MIT

package paths

import (
	"log"
	"os"
	"path/filepath"
)

func init() {
	// On iOS, os.UserHomeDir() returns the app container root.
	// Writing directly to the app container root is blocked by the iOS sandbox,
	// which causes immediate crashes on startup when trying to write .renbrowser or .reticulum-go.
	// We override the HOME environment variable to point to the app's Documents directory,
	// which is fully writable and allowed by the sandbox.
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("paths init: failed to get user home dir: %v", err)
		return
	}
	docDir := filepath.Join(home, "Documents")
	if err := os.Setenv("HOME", docDir); err != nil {
		log.Printf("paths init: failed to set HOME env: %v", err)
		return
	}
	SetDataRoot(docDir)
}

func InitAndroid() {}

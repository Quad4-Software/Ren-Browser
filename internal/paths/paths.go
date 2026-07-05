// SPDX-License-Identifier: MIT
package paths

import (
	"os"
	"path/filepath"
)

var dataRoot string

func SetDataRoot(root string) {
	dataRoot = root
}

func DataRoot() string {
	if dataRoot != "" {
		return dataRoot
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}

func Join(elem ...string) string {
	return filepath.Join(append([]string{DataRoot()}, elem...)...)
}

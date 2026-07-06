//go:build !android

// SPDX-License-Identifier: MIT

package app

import (
	"os"
	"path/filepath"
)

func shouldResetStoredDownloadDir(string) bool {
	return false
}

func writeDownloadBytes(dir, name string, data []byte) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	dest := uniqueFilePath(filepath.Join(dir, name))
	if err := os.WriteFile(dest, data, 0o600); err != nil {
		return "", err
	}
	return dest, nil
}

func platformOpenPath(path string) error {
	return openPathDesktop(path)
}

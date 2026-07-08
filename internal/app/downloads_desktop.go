//go:build !android

// SPDX-License-Identifier: MIT

package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"renbrowser/internal/paths"
)

func shouldResetStoredDownloadDir(dir string) bool {
	if runtime.GOOS != "ios" {
		return false
	}
	return iosDownloadDirNeedsReset(dir, paths.DataRoot())
}

// iosDownloadDirNeedsReset reports whether a persisted download path is outside
// the writable Documents data root (e.g. container-root Downloads).
func iosDownloadDirNeedsReset(dir, root string) bool {
	dir = filepath.Clean(strings.TrimSpace(dir))
	if dir == "" || isRootLevelDownloadDir(dir) || isTempDownloadDir(dir) {
		return true
	}
	root = filepath.Clean(strings.TrimSpace(root))
	if root == "" || root == "." {
		return false
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return true
	}
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
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

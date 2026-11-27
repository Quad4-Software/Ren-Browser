//go:build !android

// SPDX-License-Identifier: MIT

package app

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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
	name = sanitizeDownloadFilename(name)
	for i := 0; i < 1000; i++ {
		candidate := name
		if i != 0 {
			ext := filepath.Ext(name)
			base := strings.TrimSuffix(name, ext)
			candidate = base + " (" + strconv.Itoa(i) + ")" + ext
		}
		dest := filepath.Join(dir, candidate)
		file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- confined by sanitized leaf name
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, writeErr := file.Write(data)
		closeErr := file.Close()
		if writeErr != nil {
			_ = os.Remove(dest)
			return "", writeErr
		}
		if closeErr != nil {
			_ = os.Remove(dest)
			return "", closeErr
		}
		return dest, nil
	}
	return "", os.ErrExist
}

func platformOpenPath(path string) error {
	return openPathDesktop(path)
}

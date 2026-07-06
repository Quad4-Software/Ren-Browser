//go:build android

// SPDX-License-Identifier: MIT

package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"

	"renbrowser/internal/paths"
)

func shouldResetStoredDownloadDir(dir string) bool {
	dir = filepath.Clean(strings.TrimSpace(dir))
	if dir == "" {
		return true
	}
	if isRootLevelDownloadDir(dir) {
		return true
	}
	root := filepath.Clean(paths.DataRoot())
	if root == "" || root == "." {
		return false
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func writeDownloadBytes(dir, name string, data []byte) (string, error) {
	_ = dir
	name = sanitizeDownloadFilename(name)
	if name == "" {
		return "", errors.New("invalid download filename")
	}
	temp, err := os.CreateTemp("", "renbrowser-dl-*")
	if err != nil {
		return "", err
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return "", err
	}
	if err := temp.Close(); err != nil {
		return "", err
	}
	payload, err := json.Marshal(map[string]string{
		"name": name,
		"path": tempPath,
	})
	if err != nil {
		return "", err
	}
	dest, ok := androidBridgeStringString("commitDownloadFileJson", string(payload))
	if !ok || strings.TrimSpace(dest) == "" {
		return "", errors.New("failed to save download to the Downloads folder")
	}
	return filepath.Clean(dest), nil
}

func platformOpenPath(path string) error {
	uri := path
	if !strings.Contains(path, "://") {
		uri = "file://" + path
	}
	application.Android.OpenURL(uri)
	return nil
}

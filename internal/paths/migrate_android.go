//go:build android

// SPDX-License-Identifier: MIT

package paths

import (
	"os"
	"path/filepath"
)

var androidDataDirs = []string{".renbrowser", ".reticulum-go"}

func migrateAndroidStorage(legacy, root string) {
	if legacy == "" || root == "" {
		return
	}
	legacy = filepath.Clean(legacy)
	root = filepath.Clean(root)
	if legacy == root {
		return
	}
	for _, name := range androidDataDirs {
		migrateAndroidDir(filepath.Join(legacy, name), filepath.Join(root, name))
	}
}

func migrateAndroidDir(src, dst string) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	if src == dst {
		return
	}
	srcInfo, err := os.Stat(src)
	if err != nil || !srcInfo.IsDir() {
		return
	}
	if _, err := os.Stat(dst); err == nil {
		return
	}
	if err := copyAndroidTree(src, dst); err != nil {
		return
	}
	_ = os.RemoveAll(src)
}

func copyAndroidTree(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		data, err := os.ReadFile(path) //nolint:gosec // one-time migration from legacy app storage
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600) //nolint:gosec // migration target under app data root
	})
}

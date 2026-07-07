// SPDX-License-Identifier: MIT
package plugins

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func DirSizeBytes(dir string) (int64, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return 0, nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return info.Size(), nil
	}
	var total int64
	err = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		fileInfo, err := d.Info()
		if err != nil {
			return err
		}
		total += fileInfo.Size()
		return nil
	})
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SPDX-License-Identifier: MIT
package db

import (
	"fmt"
	"os"
	"path/filepath"

	"renbrowser/internal/apperrors"
)

type HealthResult struct {
	OK     bool
	Detail string
}

func (d *DB) QuickCheck() (HealthResult, error) {
	if d == nil || d.sql == nil {
		return HealthResult{}, fmt.Errorf("database not open")
	}
	var result string
	if err := d.sql.QueryRow(`PRAGMA quick_check`).Scan(&result); err != nil {
		return HealthResult{}, err
	}
	if result == "ok" {
		return HealthResult{OK: true}, nil
	}
	return HealthResult{OK: false, Detail: result}, nil
}

func IsCorruptError(err error) bool {
	return apperrors.IsCorruptError(err)
}

func RemoveFiles(path string) error {
	path = filepath.Clean(path)
	for _, target := range []string{path, path + "-wal", path + "-shm"} {
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

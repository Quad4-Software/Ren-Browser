// SPDX-License-Identifier: MIT
package db

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const backupSuffix = ".bak"

func BackupDir(dbPath string) string {
	return filepath.Join(filepath.Dir(dbPath), "backups")
}

func backupName(dbPath string, ts time.Time) string {
	base := filepath.Base(dbPath)
	return base + "." + ts.UTC().Format("20060102-150405") + backupSuffix
}

func (d *DB) BackupTo(destPath string) error {
	if d == nil || d.sql == nil {
		return fmt.Errorf("database not open")
	}
	destPath = filepath.Clean(destPath)
	if destPath == "" || destPath == "." {
		return fmt.Errorf("backup path is required")
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o700); err != nil {
		return err
	}
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("backup destination already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := d.Checkpoint(); err != nil {
		return fmt.Errorf("checkpoint: %w", err)
	}
	escaped := strings.ReplaceAll(destPath, "'", "''")
	if _, err := d.sql.Exec(`VACUUM INTO '` + escaped + `'`); err != nil { // #nosec G201,G202 -- path escaped, not user SQL
		return fmt.Errorf("vacuum into: %w", err)
	}
	return nil
}

func (d *DB) BackupBesideDB() (string, error) {
	if d == nil || d.path == "" {
		return "", fmt.Errorf("database path is not set")
	}
	ts := time.Now().UTC()
	dest := filepath.Join(BackupDir(d.path), backupName(d.path, ts))
	if err := d.BackupTo(dest); err != nil {
		return "", err
	}
	return dest, nil
}

func LatestBackup(dbPath string) (string, error) {
	dir := BackupDir(dbPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	base := filepath.Base(dbPath)
	prefix := base + "."
	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, backupSuffix) {
			matches = append(matches, filepath.Join(dir, name))
		}
	}
	if len(matches) == 0 {
		return "", nil
	}
	sort.Strings(matches)
	return matches[len(matches)-1], nil
}

func RestoreBackup(dbPath, backupPath string) error {
	dbPath = filepath.Clean(dbPath)
	backupPath = filepath.Clean(backupPath)
	if dbPath == "" || backupPath == "" {
		return fmt.Errorf("paths are required")
	}
	info, err := os.Stat(backupPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("backup is a directory")
	}
	if err := RemoveFiles(dbPath); err != nil {
		return err
	}
	if err := copyFile(backupPath, dbPath); err != nil {
		return err
	}
	return os.Chmod(dbPath, 0o600)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) // #nosec G304 -- caller-controlled backup path
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- caller-controlled backup path
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

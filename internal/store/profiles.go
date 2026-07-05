// SPDX-License-Identifier: MIT
package store

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"renbrowser/internal/brand"
)

var profileNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`)

func SanitizeProfileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "default" {
		return "default"
	}
	if !profileNamePattern.MatchString(name) {
		return "default"
	}
	return name
}

func ProfilePath(name string) string {
	name = SanitizeProfileName(name)
	return brand.ProfileDBPath(name)
}

func ProfilesDir() string {
	return brand.ProfilesDir()
}

func ListProfileNames() ([]string, error) {
	names := []string{"default"}
	dir := ProfilesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return names, nil
		}
		return nil, err
	}
	seen := map[string]bool{"default": true}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !profileNamePattern.MatchString(name) {
			continue
		}
		dbPath := filepath.Join(dir, name, brand.DBFileName)
		if _, err := os.Stat(dbPath); err != nil {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names, nil
}

// SPDX-License-Identifier: MIT
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"renbrowser/internal/db"
)

const tabPagesDirName = "tab-pages"

type tabBodyFile struct {
	HTML    string `json:"html,omitempty"`
	LastRaw string `json:"lastRaw,omitempty"`
}

func (s *Store) tabPagesDir() string {
	return filepath.Join(filepath.Dir(s.path), tabPagesDirName)
}

func safeTabID(id string) (string, bool) {
	if id == "" || len(id) > 80 {
		return "", false
	}
	if strings.Contains(id, "..") {
		return "", false
	}
	for _, r := range id {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", false
	}
	return id, true
}

func isEditorTabURL(url string) bool {
	switch strings.ToLower(strings.TrimSpace(url)) {
	case "editor", "editor:":
		return true
	default:
		return false
	}
}

func (s *Store) tabBodyPath(id string) (string, bool) {
	safe, ok := safeTabID(id)
	if !ok {
		return "", false
	}
	return filepath.Join(s.tabPagesDir(), safe+".json"), true
}

func (s *Store) writeTabBody(id, html, lastRaw string) error {
	path, ok := s.tabBodyPath(id)
	if !ok {
		return nil
	}
	if html == "" && lastRaw == "" {
		return s.removeTabBody(id)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.Marshal(tabBodyFile{HTML: html, LastRaw: lastRaw})
	if err != nil {
		return err
	}
	return atomicWriteFile(path, raw, 0o600)
}

func (s *Store) readTabBody(id string) (html, lastRaw string, ok bool) {
	path, valid := s.tabBodyPath(id)
	if !valid {
		return "", "", false
	}
	raw, err := os.ReadFile(path) // #nosec G304 -- path is tab-pages/<sanitized-id>.json under the profile dir
	if err != nil {
		return "", "", false
	}
	var body tabBodyFile
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", "", false
	}
	if body.HTML == "" && body.LastRaw == "" {
		return "", "", false
	}
	return body.HTML, body.LastRaw, true
}

func (s *Store) removeTabBody(id string) error {
	path, ok := s.tabBodyPath(id)
	if !ok {
		return nil
	}
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *Store) pruneTabBodies(keep map[string]struct{}) {
	dir := s.tabPagesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if !strings.HasSuffix(name, ".json") {
			_ = os.Remove(filepath.Join(dir, name))
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if _, ok := keep[id]; ok {
			continue
		}
		_ = os.Remove(filepath.Join(dir, name))
	}
}

func (s *Store) clearTabBodies() {
	_ = os.RemoveAll(s.tabPagesDir())
}

func (s *Store) hydrateTabBodies(tabs []TabSnapshot, includeAll bool) []TabSnapshot {
	if len(tabs) == 0 {
		return tabs
	}
	out := make([]TabSnapshot, len(tabs))
	copy(out, tabs)
	for i, tab := range out {
		if !includeAll && !tab.Active && !isEditorTabURL(tab.URL) {
			continue
		}
		html, lastRaw, ok := s.readTabBody(tab.ID)
		if !ok {
			continue
		}
		out[i].HTML = html
		out[i].LastRaw = lastRaw
	}
	return out
}

func (s *Store) migrateTabBodiesFromSQLite() error {
	if s.db == nil {
		return nil
	}
	rows, err := s.db.Tabs()
	if err != nil {
		return err
	}
	moved := false
	for _, row := range rows {
		if row.HTML == "" && row.LastRaw == "" {
			continue
		}
		if err := s.writeTabBody(row.ID, row.HTML, row.LastRaw); err != nil {
			return err
		}
		moved = true
	}
	if !moved {
		return nil
	}
	cleared := make([]db.TabRow, len(rows))
	for i, row := range rows {
		row.HTML = ""
		row.LastRaw = ""
		cleared[i] = row
	}
	return s.db.SaveTabs(cleared)
}

func stripTabBodies(tabs []TabSnapshot) []TabSnapshot {
	out := make([]TabSnapshot, len(tabs))
	for i, tab := range tabs {
		out[i] = tab
		out[i].HTML = ""
		out[i].LastRaw = ""
	}
	return out
}

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	ok = true
	return nil
}

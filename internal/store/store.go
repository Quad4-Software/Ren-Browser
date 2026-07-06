// SPDX-License-Identifier: MIT
package store

import (
	"encoding/json"
	"os"
	"sync"

	"renbrowser/internal/apperrors"
	"renbrowser/internal/brand"
	"renbrowser/internal/db"
	"renbrowser/internal/nomadnet"
)

type TabSnapshot struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Active      bool   `json:"active"`
	Pinned      bool   `json:"pinned,omitempty"`
	HTML        string `json:"html,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Error       string `json:"error,omitempty"`
	DurationMs  int64  `json:"durationMs,omitempty"`
	LastRaw     string `json:"lastRaw,omitempty"`
	PageFG      string `json:"pageFg,omitempty"`
	PageBG      string `json:"pageBg,omitempty"`
}

type HistoryEntry struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	NodeHash  string `json:"nodeHash"`
	VisitedAt int64  `json:"visitedAt"`
}

type Store struct {
	mu            sync.RWMutex
	path          string
	db            *db.DB
	corrupt       bool
	corruptDetail string
}

func DefaultPath() string {
	return db.DefaultPath()
}

func legacyJSONPath() string {
	return brand.LegacyStatePath()
}

func Open(path string) (*Store, error) {
	if path == "" {
		path = DefaultPath()
	}
	database, err := db.Open(path)
	if err != nil {
		if db.IsCorruptError(err) {
			return &Store{path: path, corrupt: true, corruptDetail: err.Error()}, nil
		}
		return nil, err
	}
	s := &Store{path: path, db: database}
	if health := s.healthLocked(); !health.OK && health.Kind == string(apperrors.KindDatabaseCorrupt) {
		s.corrupt = true
		s.corruptDetail = health.Detail
	}
	if err := s.migrateLegacyJSON(); err != nil {
		_ = database.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) DB() *db.DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db
}

type legacyData struct {
	Favorites []string      `json:"favorites"`
	Recent    []string      `json:"recent"`
	Tabs      []TabSnapshot `json:"tabs"`
}

func (s *Store) migrateLegacyJSON() error {
	if s.path != DefaultPath() {
		return nil
	}
	jsonPath := legacyJSONPath()
	raw, err := os.ReadFile(jsonPath) // #nosec G304 -- fixed path under user home .renbrowser
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}

	var legacy legacyData
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return err
	}

	favCount, _ := s.db.Count("favorites")
	if favCount == 0 {
		for _, url := range legacy.Favorites {
			_, _ = s.db.AddFavorite(url)
		}
	}

	histCount, _ := s.db.Count("history")
	if histCount == 0 {
		for _, url := range legacy.Recent {
			_ = s.AddHistory(url, "", "")
		}
	}

	tabCount, _ := s.db.Count("tabs")
	if tabCount == 0 && len(legacy.Tabs) > 0 {
		_ = s.SaveTabs(legacy.Tabs)
	}
	return nil
}

func (s *Store) UpsertNode(node nomadnet.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.UpsertNode(db.NodeRow{
		Hash:      node.Hash,
		Name:      node.Name,
		Hops:      node.Hops,
		Enabled:   node.Enabled,
		Timestamp: node.Timestamp,
		MaxSizeKB: node.MaxSizeKB,
		LastSeen:  node.LastSeen,
	})
}

func (s *Store) ListNodes() ([]nomadnet.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.ListNodes()
	if err != nil {
		return nil, err
	}
	out := make([]nomadnet.Node, len(rows))
	for i, r := range rows {
		out[i] = nomadnet.Node{
			Hash:      r.Hash,
			Name:      r.Name,
			Hops:      r.Hops,
			Enabled:   r.Enabled,
			Timestamp: r.Timestamp,
			MaxSizeKB: r.MaxSizeKB,
			LastSeen:  r.LastSeen,
		}
	}
	return out, nil
}

func (s *Store) AddHistory(url, title, nodeHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.AddHistory(url, title, nodeHash, 0)
}

func (s *Store) BrowsingHistory(limit int) ([]HistoryEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.ListHistory(limit)
	if err != nil {
		return nil, err
	}
	out := make([]HistoryEntry, len(rows))
	for i, r := range rows {
		out[i] = HistoryEntry{
			ID:        r.ID,
			URL:       r.URL,
			Title:     r.Title,
			NodeHash:  r.NodeHash,
			VisitedAt: r.VisitedAt,
		}
	}
	return out, nil
}

func (s *Store) ClearBrowsingHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.ClearHistory()
}

func (s *Store) Favorites() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out, err := s.db.Favorites()
	if err != nil {
		return []string{}
	}
	return out
}

func (s *Store) AddFavorite(url string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out, err := s.db.AddFavorite(url)
	if err != nil {
		return []string{}
	}
	return out
}

func (s *Store) RemoveFavorite(url string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out, err := s.db.RemoveFavorite(url)
	if err != nil {
		return []string{}
	}
	return out
}

func (s *Store) PushRecent(url string) []string {
	_ = s.AddHistory(url, "", "")
	rows, err := s.BrowsingHistory(50)
	if err != nil {
		return []string{url}
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.URL
	}
	return out
}

func (s *Store) Recent() []string {
	rows, err := s.BrowsingHistory(50)
	if err != nil {
		return []string{}
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.URL
	}
	return out
}

func (s *Store) Tabs() []TabSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.Tabs()
	if err != nil {
		return []TabSnapshot{}
	}
	return tabRowsToSnapshots(rows)
}

func (s *Store) SaveTabs(tabs []TabSnapshot) []TabSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return tabs
	}
	tabs = clampTabSnapshots(tabs)
	rows := snapshotsToTabRows(tabs)
	if err := s.db.SaveTabs(rows); err != nil {
		s.noteWriteError(err)
		return tabs
	}
	return tabs
}

func (s *Store) GetSetting(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.GetSetting(key)
}

func (s *Store) SetSetting(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return os.ErrInvalid
	}
	err := s.db.SetSetting(key, value)
	s.noteWriteError(err)
	return err
}

func tabRowsToSnapshots(rows []db.TabRow) []TabSnapshot {
	out := make([]TabSnapshot, len(rows))
	for i, r := range rows {
		out[i] = TabSnapshot{
			ID:          r.ID,
			Title:       r.Title,
			URL:         r.URL,
			Active:      r.Active,
			Pinned:      r.Pinned,
			HTML:        r.HTML,
			ContentType: r.ContentType,
			Error:       r.Error,
			DurationMs:  r.DurationMs,
			LastRaw:     r.LastRaw,
			PageFG:      r.PageFG,
			PageBG:      r.PageBG,
		}
	}
	return out
}

func snapshotsToTabRows(tabs []TabSnapshot) []db.TabRow {
	out := make([]db.TabRow, len(tabs))
	for i, t := range tabs {
		out[i] = db.TabRow{
			ID:          t.ID,
			Title:       t.Title,
			URL:         t.URL,
			Active:      t.Active,
			Pinned:      t.Pinned,
			HTML:        t.HTML,
			ContentType: t.ContentType,
			Error:       t.Error,
			DurationMs:  t.DurationMs,
			LastRaw:     t.LastRaw,
			PageFG:      t.PageFG,
			PageBG:      t.PageBG,
		}
	}
	return out
}

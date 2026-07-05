package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"renbrowser/internal/paths"
)

type DB struct {
	sql *sql.DB
}

func DefaultPath() string {
	return paths.Join(".renbrowser", "renbrowser.db")
}

func Open(path string) (*DB, error) {
	if path == "" {
		path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)

	d := &DB{sql: conn}
	if err := d.migrate(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) Close() error {
	if d == nil || d.sql == nil {
		return nil
	}
	return d.sql.Close()
}

func (d *DB) migrate() error {
	if _, err := d.sql.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return fmt.Errorf("wal: %w", err)
	}
	if _, err := d.sql.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		return fmt.Errorf("foreign_keys: %w", err)
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS nodes (
			hash TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			hops INTEGER NOT NULL DEFAULT 0,
			enabled INTEGER NOT NULL DEFAULT 1,
			timestamp INTEGER NOT NULL DEFAULT 0,
			max_size_kb INTEGER NOT NULL DEFAULT 0,
			last_seen INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			node_hash TEXT NOT NULL DEFAULT '',
			visited_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_history_visited ON history(visited_at DESC)`,
		`CREATE TABLE IF NOT EXISTS favorites (
			url TEXT PRIMARY KEY,
			added_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tabs (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			url TEXT NOT NULL,
			active INTEGER NOT NULL DEFAULT 0,
			pinned INTEGER NOT NULL DEFAULT 0,
			html TEXT NOT NULL DEFAULT '',
			content_type TEXT NOT NULL DEFAULT '',
			error TEXT NOT NULL DEFAULT '',
			duration_ms INTEGER NOT NULL DEFAULT 0,
			last_raw TEXT NOT NULL DEFAULT '',
			page_fg TEXT NOT NULL DEFAULT '',
			page_bg TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		if _, err := d.sql.Exec(stmt); err != nil {
			return err
		}
	}
	if _, err := d.sql.Exec(`ALTER TABLE tabs ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}

type NodeRow struct {
	Hash      string
	Name      string
	Hops      uint8
	Enabled   bool
	Timestamp int64
	MaxSizeKB int16
	LastSeen  int64
}

func (d *DB) UpsertNode(n NodeRow) error {
	_, err := d.sql.Exec(
		`INSERT INTO nodes (hash, name, hops, enabled, timestamp, max_size_kb, last_seen)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(hash) DO UPDATE SET
		   name=excluded.name,
		   hops=excluded.hops,
		   enabled=excluded.enabled,
		   timestamp=excluded.timestamp,
		   max_size_kb=excluded.max_size_kb,
		   last_seen=excluded.last_seen`,
		n.Hash, n.Name, n.Hops, boolInt(n.Enabled), n.Timestamp, n.MaxSizeKB, n.LastSeen,
	)
	return err
}

func (d *DB) ListNodes() ([]NodeRow, error) {
	rows, err := d.sql.Query(
		`SELECT hash, name, hops, enabled, timestamp, max_size_kb, last_seen
		 FROM nodes ORDER BY last_seen DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []NodeRow
	for rows.Next() {
		var n NodeRow
		var enabled int
		if err := rows.Scan(&n.Hash, &n.Name, &n.Hops, &enabled, &n.Timestamp, &n.MaxSizeKB, &n.LastSeen); err != nil {
			return nil, err
		}
		n.Enabled = enabled != 0
		out = append(out, n)
	}
	return out, rows.Err()
}

type HistoryRow struct {
	ID        int64
	URL       string
	Title     string
	NodeHash  string
	VisitedAt int64
}

func (d *DB) AddHistory(url, title, nodeHash string, visitedAt int64) error {
	if visitedAt == 0 {
		visitedAt = time.Now().Unix()
	}
	_, err := d.sql.Exec(
		`INSERT INTO history (url, title, node_hash, visited_at) VALUES (?, ?, ?, ?)`,
		url, title, nodeHash, visitedAt,
	)
	return err
}

func (d *DB) ListHistory(limit int) ([]HistoryRow, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := d.sql.Query(
		`SELECT id, url, title, node_hash, visited_at FROM history
		 ORDER BY visited_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []HistoryRow
	for rows.Next() {
		var h HistoryRow
		if err := rows.Scan(&h.ID, &h.URL, &h.Title, &h.NodeHash, &h.VisitedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (d *DB) ListRecentURLs(limit int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := d.sql.Query(
		`SELECT url FROM history ORDER BY visited_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		out = append(out, url)
	}
	return out, rows.Err()
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

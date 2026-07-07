// SPDX-License-Identifier: MIT
package db

import (
	"fmt"
	"strings"
	"time"
)

func (d *DB) Favorites() ([]string, error) {
	rows, err := d.sql.Query(`SELECT url FROM favorites ORDER BY added_at ASC`)
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

func (d *DB) AddFavorite(url string) ([]string, error) {
	_, err := d.sql.Exec(
		`INSERT INTO favorites (url, added_at) VALUES (?, ?)
		 ON CONFLICT(url) DO NOTHING`,
		url, time.Now().Unix(),
	)
	if err != nil {
		return nil, err
	}
	return d.Favorites()
}

func (d *DB) RemoveFavorite(url string) ([]string, error) {
	_, err := d.sql.Exec(`DELETE FROM favorites WHERE url = ?`, url)
	if err != nil {
		return nil, err
	}
	return d.Favorites()
}

type TabRow struct {
	ID          string
	Title       string
	URL         string
	Active      bool
	Pinned      bool
	HTML        string
	ContentType string
	Error       string
	DurationMs  int64
	LastRaw     string
	PageFG      string
	PageBG      string
	SortOrder   int
}

func (d *DB) Tabs() ([]TabRow, error) {
	rows, err := d.sql.Query(
		`SELECT id, title, url, active, pinned, html, content_type, error, duration_ms, last_raw, page_fg, page_bg, sort_order
		 FROM tabs ORDER BY sort_order ASC, rowid ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TabRow
	for rows.Next() {
		var t TabRow
		var active int
		var pinned int
		if err := rows.Scan(
			&t.ID, &t.Title, &t.URL, &active, &pinned,
			&t.HTML, &t.ContentType, &t.Error, &t.DurationMs, &t.LastRaw,
			&t.PageFG, &t.PageBG, &t.SortOrder,
		); err != nil {
			return nil, err
		}
		t.Active = active != 0
		t.Pinned = pinned != 0
		out = append(out, t)
	}
	return out, rows.Err()
}

func (d *DB) SaveTabs(tabs []TabRow) error {
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(
		`INSERT INTO tabs (
			id, title, url, active, pinned, html, content_type, error, duration_ms, last_raw, page_fg, page_bg, sort_order
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title=excluded.title,
			url=excluded.url,
			active=excluded.active,
			pinned=excluded.pinned,
			html=excluded.html,
			content_type=excluded.content_type,
			error=excluded.error,
			duration_ms=excluded.duration_ms,
			last_raw=excluded.last_raw,
			page_fg=excluded.page_fg,
			page_bg=excluded.page_bg,
			sort_order=excluded.sort_order`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	keepIDs := make([]any, len(tabs))
	for i, t := range tabs {
		if _, err := stmt.Exec(
			t.ID, t.Title, t.URL, boolInt(t.Active), boolInt(t.Pinned),
			t.HTML, t.ContentType, t.Error, t.DurationMs, t.LastRaw,
			t.PageFG, t.PageBG, i,
		); err != nil {
			return err
		}
		if saveTabsAfterInsertHook != nil {
			if err := saveTabsAfterInsertHook(i + 1); err != nil {
				return err
			}
		}
		keepIDs[i] = t.ID
	}

	if len(keepIDs) == 0 {
		if _, err := tx.Exec(`DELETE FROM tabs`); err != nil {
			return err
		}
	} else {
		placeholders := make([]string, len(keepIDs))
		for i := range keepIDs {
			placeholders[i] = "?"
		}
		query := `DELETE FROM tabs WHERE id NOT IN (` + strings.Join(placeholders, ",") + `)`
		if _, err := tx.Exec(query, keepIDs...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *DB) Count(table string) (int, error) {
	switch table {
	case "favorites", "history", "tabs", "nodes", "settings":
	default:
		return 0, fmt.Errorf("invalid table %q", table)
	}
	var n int
	err := d.sql.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n) // #nosec G202 -- table name allowlisted above
	return n, err
}

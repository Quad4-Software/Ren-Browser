// SPDX-License-Identifier: MIT
package db

import (
	"database/sql"
	"errors"
)

type PluginRow struct {
	ID           string
	Enabled      bool
	SettingsJSON string
}

func (d *DB) UpsertPlugin(id string, enabled bool, settingsJSON string) error {
	if settingsJSON == "" {
		settingsJSON = "{}"
	}
	_, err := d.sql.Exec(
		`INSERT INTO plugins (id, enabled, settings_json) VALUES (?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET enabled = excluded.enabled, settings_json = excluded.settings_json`,
		id, boolInt(enabled), settingsJSON,
	)
	return err
}

func (d *DB) GetPlugin(id string) (PluginRow, error) {
	var row PluginRow
	var enabled int
	err := d.sql.QueryRow(
		`SELECT id, enabled, settings_json FROM plugins WHERE id = ?`, id,
	).Scan(&row.ID, &enabled, &row.SettingsJSON)
	if err != nil {
		return PluginRow{}, err
	}
	row.Enabled = enabled != 0
	return row, nil
}

func (d *DB) ListPlugins() ([]PluginRow, error) {
	rows, err := d.sql.Query(`SELECT id, enabled, settings_json FROM plugins ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PluginRow
	for rows.Next() {
		var row PluginRow
		var enabled int
		if err := rows.Scan(&row.ID, &enabled, &row.SettingsJSON); err != nil {
			return nil, err
		}
		row.Enabled = enabled != 0
		out = append(out, row)
	}
	return out, rows.Err()
}

func (d *DB) SetPluginEnabled(id string, enabled bool) error {
	_, err := d.sql.Exec(`UPDATE plugins SET enabled = ? WHERE id = ?`, boolInt(enabled), id)
	return err
}

func (d *DB) DeletePlugin(id string) error {
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM plugin_settings WHERE plugin_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM plugins WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *DB) GetPluginSetting(pluginID, key string) (string, error) {
	var value string
	err := d.sql.QueryRow(
		`SELECT value FROM plugin_settings WHERE plugin_id = ? AND key = ?`,
		pluginID, key,
	).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (d *DB) SetPluginSetting(pluginID, key, value string) error {
	_, err := d.sql.Exec(
		`INSERT INTO plugin_settings (plugin_id, key, value) VALUES (?, ?, ?)
		 ON CONFLICT(plugin_id, key) DO UPDATE SET value = excluded.value`,
		pluginID, key, value,
	)
	return err
}

func (d *DB) DeletePluginSettings(pluginID string) error {
	_, err := d.sql.Exec(`DELETE FROM plugin_settings WHERE plugin_id = ?`, pluginID)
	return err
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

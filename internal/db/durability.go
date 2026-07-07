// SPDX-License-Identifier: MIT
package db

import (
	"fmt"
)

const schemaVersion = 1

// saveTabsAfterInsertHook is set by reliability tests to simulate mid-write failures.
var saveTabsAfterInsertHook func(inserted int) error

func (d *DB) configureDurability() error {
	pragmas := []struct {
		name string
		sql  string
	}{
		{"journal_mode", `PRAGMA journal_mode=WAL`},
		{"foreign_keys", `PRAGMA foreign_keys=ON`},
		{"busy_timeout", `PRAGMA busy_timeout=5000`},
		{"synchronous", `PRAGMA synchronous=FULL`},
		{"wal_autocheckpoint", `PRAGMA wal_autocheckpoint=1000`},
	}
	for _, p := range pragmas {
		if _, err := d.sql.Exec(p.sql); err != nil {
			return fmt.Errorf("%s: %w", p.name, err)
		}
	}
	return nil
}

func (d *DB) Checkpoint() error {
	if d == nil || d.sql == nil {
		return nil
	}
	_, err := d.sql.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
	return err
}

func (d *DB) PragmaInt(name string) (int, error) {
	if d == nil || d.sql == nil {
		return 0, fmt.Errorf("database not open")
	}
	var value int
	err := d.sql.QueryRow(`PRAGMA ` + name).Scan(&value) // #nosec G202 -- name from test allowlist
	return value, err
}

func (d *DB) PragmaString(name string) (string, error) {
	if d == nil || d.sql == nil {
		return "", fmt.Errorf("database not open")
	}
	var value string
	err := d.sql.QueryRow(`PRAGMA ` + name).Scan(&value) // #nosec G202 -- name from test allowlist
	return value, err
}

func (d *DB) SchemaVersion() (int, error) {
	return d.PragmaInt("user_version")
}

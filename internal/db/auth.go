// SPDX-License-Identifier: MIT
package db

import (
	"database/sql"
	"errors"
	"time"
)

const (
	AuthPepperSettingKey = "auth_pepper"
)

type AuthCredential struct {
	PasswordHash string
	UpdatedAt    int64
}

type AuthBruteState struct {
	ClientHash  string
	FailCount   int
	BannedUntil int64
	Trusted     bool
}

type AuthSession struct {
	TokenHash string
	CreatedAt int64
	ExpiresAt int64
}

func (d *DB) authMigrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS auth_credentials (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			password_hash TEXT NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS auth_brute_state (
			client_hash TEXT PRIMARY KEY,
			fail_count INTEGER NOT NULL DEFAULT 0,
			banned_until INTEGER NOT NULL DEFAULT 0,
			trusted INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS auth_sessions (
			token_hash TEXT PRIMARY KEY,
			created_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires ON auth_sessions(expires_at)`,
	}
	for _, stmt := range stmts {
		if _, err := d.sql.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) AuthEnabled() (bool, error) {
	cred, err := d.GetAuthCredential()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return cred.PasswordHash != "", nil
}

func (d *DB) GetAuthCredential() (AuthCredential, error) {
	var cred AuthCredential
	err := d.sql.QueryRow(
		`SELECT password_hash, updated_at FROM auth_credentials WHERE id = 1`,
	).Scan(&cred.PasswordHash, &cred.UpdatedAt)
	return cred, err
}

func (d *DB) SetAuthCredential(hash string) error {
	now := time.Now().Unix()
	_, err := d.sql.Exec(
		`INSERT INTO auth_credentials (id, password_hash, updated_at) VALUES (1, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET password_hash = excluded.password_hash, updated_at = excluded.updated_at`,
		hash, now,
	)
	return err
}

func (d *DB) ClearAuthCredential() error {
	_, err := d.sql.Exec(`DELETE FROM auth_credentials WHERE id = 1`)
	return err
}

func (d *DB) ClearAuthSessions() error {
	_, err := d.sql.Exec(`DELETE FROM auth_sessions`)
	return err
}

func (d *DB) ClearAuthBruteState() error {
	_, err := d.sql.Exec(`DELETE FROM auth_brute_state`)
	return err
}

func (d *DB) GetAuthBruteState(clientHash string) (AuthBruteState, error) {
	var state AuthBruteState
	var trusted int
	err := d.sql.QueryRow(
		`SELECT client_hash, fail_count, banned_until, trusted FROM auth_brute_state WHERE client_hash = ?`,
		clientHash,
	).Scan(&state.ClientHash, &state.FailCount, &state.BannedUntil, &trusted)
	if err != nil {
		return state, err
	}
	state.Trusted = trusted != 0
	return state, nil
}

func (d *DB) UpsertAuthBruteState(state AuthBruteState) error {
	_, err := d.sql.Exec(
		`INSERT INTO auth_brute_state (client_hash, fail_count, banned_until, trusted)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(client_hash) DO UPDATE SET
		   fail_count = excluded.fail_count,
		   banned_until = excluded.banned_until,
		   trusted = excluded.trusted`,
		state.ClientHash, state.FailCount, state.BannedUntil, boolInt(state.Trusted),
	)
	return err
}

func (d *DB) MarkAuthClientTrusted(clientHash string) error {
	_, err := d.sql.Exec(
		`INSERT INTO auth_brute_state (client_hash, fail_count, banned_until, trusted)
		 VALUES (?, 0, 0, 1)
		 ON CONFLICT(client_hash) DO UPDATE SET trusted = 1, fail_count = 0, banned_until = 0`,
		clientHash,
	)
	return err
}

func (d *DB) CreateAuthSession(tokenHash string, expiresAt int64) error {
	now := time.Now().Unix()
	_, err := d.sql.Exec(
		`INSERT INTO auth_sessions (token_hash, created_at, expires_at) VALUES (?, ?, ?)`,
		tokenHash, now, expiresAt,
	)
	return err
}

func (d *DB) GetAuthSession(tokenHash string) (AuthSession, error) {
	var session AuthSession
	err := d.sql.QueryRow(
		`SELECT token_hash, created_at, expires_at FROM auth_sessions WHERE token_hash = ?`,
		tokenHash,
	).Scan(&session.TokenHash, &session.CreatedAt, &session.ExpiresAt)
	return session, err
}

func (d *DB) DeleteAuthSession(tokenHash string) error {
	_, err := d.sql.Exec(`DELETE FROM auth_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

func (d *DB) PruneExpiredAuthSessions() error {
	now := time.Now().Unix()
	_, err := d.sql.Exec(`DELETE FROM auth_sessions WHERE expires_at <= ?`, now)
	return err
}

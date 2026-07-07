// SPDX-License-Identifier: MIT
package store

import (
	"os"
	"renbrowser/internal/apperrors"
	"renbrowser/internal/db"
)

type Health struct {
	OK     bool   `json:"ok"`
	Kind   string `json:"kind,omitempty"`
	Detail string `json:"detail,omitempty"`
	Path   string `json:"path"`
}

func (s *Store) Health() Health {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.healthLocked()
}

func (s *Store) healthLocked() Health {
	out := Health{Path: s.path}
	if s.corrupt {
		out.OK = false
		out.Kind = string(apperrors.KindDatabaseCorrupt)
		out.Detail = s.corruptDetail
		return out
	}
	if s.db == nil {
		out.OK = false
		out.Kind = string(apperrors.KindDatabaseCorrupt)
		out.Detail = "database is not open"
		return out
	}
	check, err := s.db.QuickCheck()
	if err != nil {
		out.OK = false
		if kind := apperrors.ClassifyStorage(err); kind != "" {
			out.Kind = string(kind)
		} else if apperrors.IsCorruptError(err) {
			out.Kind = string(apperrors.KindDatabaseCorrupt)
		} else {
			out.Kind = string(apperrors.KindInternal)
		}
		out.Detail = err.Error()
		return out
	}
	if !check.OK {
		out.OK = false
		out.Kind = string(apperrors.KindDatabaseCorrupt)
		out.Detail = check.Detail
		return out
	}
	out.OK = true
	return out
}

func (s *Store) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		_, _ = s.db.BackupBesideDB()
		_ = s.db.Close()
		s.db = nil
	}
	if err := db.RemoveFiles(s.path); err != nil {
		return err
	}
	database, err := db.Open(s.path)
	if err != nil {
		s.corrupt = true
		s.corruptDetail = err.Error()
		return err
	}
	s.db = database
	s.corrupt = false
	s.corruptDetail = ""
	return nil
}

func (s *Store) Backup() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return "", os.ErrInvalid
	}
	return s.db.BackupBesideDB()
}

func (s *Store) noteWriteError(err error) {
	if err == nil {
		return
	}
	if kind := apperrors.ClassifyStorage(err); kind == apperrors.KindDatabaseCorrupt {
		s.corrupt = true
		s.corruptDetail = err.Error()
	}
}

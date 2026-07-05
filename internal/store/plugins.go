// SPDX-License-Identifier: MIT
package store

import (
	"os"

	"renbrowser/internal/db"
)

func (s *Store) UpsertPlugin(id string, enabled bool, settingsJSON string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return os.ErrInvalid
	}
	return s.db.UpsertPlugin(id, enabled, settingsJSON)
}

func (s *Store) GetPlugin(id string) (db.PluginRow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return db.PluginRow{}, os.ErrInvalid
	}
	return s.db.GetPlugin(id)
}

func (s *Store) ListPlugins() ([]db.PluginRow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, os.ErrInvalid
	}
	return s.db.ListPlugins()
}

func (s *Store) SetPluginEnabled(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return os.ErrInvalid
	}
	return s.db.SetPluginEnabled(id, enabled)
}

func (s *Store) DeletePlugin(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return os.ErrInvalid
	}
	return s.db.DeletePlugin(id)
}

func (s *Store) GetPluginSetting(pluginID, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return "", os.ErrInvalid
	}
	return s.db.GetPluginSetting(pluginID, key)
}

func (s *Store) SetPluginSetting(pluginID, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return os.ErrInvalid
	}
	return s.db.SetPluginSetting(pluginID, key, value)
}

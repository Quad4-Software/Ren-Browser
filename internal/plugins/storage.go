// SPDX-License-Identifier: MIT
package plugins

import (
	"fmt"

	"renbrowser/internal/db"
)

type StorageStore interface {
	GetPluginSetting(pluginID, key string) (string, error)
	SetPluginSetting(pluginID, key, value string) error
}

type Storage struct {
	db StorageStore
}

func NewStorage(database StorageStore) *Storage {
	return &Storage{db: database}
}

func (s *Storage) Get(pluginID, key string) (string, error) {
	if s == nil || s.db == nil {
		return "", fmt.Errorf("plugin storage unavailable")
	}
	value, err := s.db.GetPluginSetting(pluginID, key)
	if err != nil {
		if db.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (s *Storage) Set(pluginID, key, value string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("plugin storage unavailable")
	}
	return s.db.SetPluginSetting(pluginID, key, value)
}

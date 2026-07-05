// SPDX-License-Identifier: MIT
package app

import (
	"renbrowser/internal/apperrors"
	"renbrowser/internal/store"
)

type StoreHealth = store.Health

func (s *BrowserService) GetStoreHealth() StoreHealth {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.publicMode {
		return StoreHealth{OK: true, Path: s.storePath}
	}
	if s.store == nil {
		return StoreHealth{
			OK:     false,
			Kind:   string(apperrors.KindDatabaseCorrupt),
			Detail: "database unavailable",
			Path:   s.storePath,
		}
	}
	return s.store.Health()
}

func (s *BrowserService) ResetDatabase() error {
	if s.publicMode {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.store == nil {
		return nil
	}
	if err := s.store.Reset(); err != nil {
		s.emitStoreHealthLocked(s.store.Health())
		return err
	}
	s.emitStoreHealthLocked(StoreHealth{OK: true, Path: s.storePath})
	return nil
}

func (s *BrowserService) emitStoreHealth() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.store == nil {
		s.emitStoreHealthLocked(StoreHealth{OK: false, Kind: string(apperrors.KindDatabaseCorrupt), Path: s.storePath})
		return
	}
	s.emitStoreHealthLocked(s.store.Health())
}

func (s *BrowserService) emitStoreHealthLocked(health StoreHealth) {
	if s.app == nil {
		return
	}
	s.app.Event.Emit("store:health", health)
}

func applyPageError(resp *PageResponse, errMsg string, body []byte) {
	if errMsg == "" {
		return
	}
	kind, detail := apperrors.ClassifyFetch(errMsg, body)
	if detail == "" {
		detail = errMsg
	}
	resp.Error = detail
	resp.ErrorKind = string(kind)
}

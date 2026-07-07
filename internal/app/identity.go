// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"fmt"

	"renbrowser/internal/rns"
)

func (s *BrowserService) ListIdentities() ([]rns.IdentityRecord, error) {
	reg, err := s.identityRegistry()
	if err != nil {
		return nil, err
	}
	return reg.List(), nil
}

func (s *BrowserService) CreateIdentity(name string) (rns.IdentityRecord, error) {
	reg, err := s.identityRegistry()
	if err != nil {
		return rns.IdentityRecord{}, err
	}
	record, err := reg.Create(name)
	if err != nil {
		return rns.IdentityRecord{}, err
	}
	s.log("info", "identity created", record.Hash)
	return record, nil
}

func (s *BrowserService) ImportIdentity(name string) (rns.IdentityRecord, error) {
	path, err := s.PickIdentityFile()
	if err != nil {
		return rns.IdentityRecord{}, err
	}
	reg, err := s.identityRegistry()
	if err != nil {
		return rns.IdentityRecord{}, err
	}
	record, err := reg.ImportFromFile(path, name)
	if err != nil {
		return rns.IdentityRecord{}, err
	}
	s.log("info", "identity imported", record.Hash)
	return record, nil
}

func (s *BrowserService) ExportIdentity(id string) error {
	reg, err := s.identityRegistry()
	if err != nil {
		return err
	}
	dest, err := s.PickIdentityExportPath()
	if err != nil {
		return err
	}
	if err := reg.Export(id, dest); err != nil {
		return err
	}
	s.log("info", "identity exported", dest)
	return nil
}

func (s *BrowserService) SetActiveIdentity(id string) (rns.IdentityRecord, error) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return rns.IdentityRecord{}, fmt.Errorf("reticulum not initialized")
	}
	if err := stack.SwitchIdentity(id); err != nil {
		return rns.IdentityRecord{}, err
	}
	reg := stack.Identities()
	if reg == nil {
		return rns.IdentityRecord{}, errors.New("identity registry not initialized")
	}
	record, err := reg.ActiveRecord()
	if err != nil {
		return rns.IdentityRecord{}, err
	}
	s.log("info", "active identity switched", record.Hash)
	if s.app != nil {
		s.app.Event.Emit("identity:changed", record)
	}
	return record, nil
}

func (s *BrowserService) RenameIdentity(id, name string) (rns.IdentityRecord, error) {
	reg, err := s.identityRegistry()
	if err != nil {
		return rns.IdentityRecord{}, err
	}
	return reg.Rename(id, name)
}

func (s *BrowserService) DeleteIdentity(id string) error {
	reg, err := s.identityRegistry()
	if err != nil {
		return err
	}
	if err := reg.Delete(id); err != nil {
		return err
	}
	s.log("info", "identity deleted", id)
	return nil
}

func (s *BrowserService) PickIdentityFile() (string, error) {
	if s.app == nil {
		return "", errors.New("file picker unavailable in server mode")
	}
	result, err := s.app.Dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		SetTitle("Select identity file").
		AddFilter("Reticulum identity", "transport_identity").
		AddFilter("All files", "*").
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *BrowserService) PickIdentityExportPath() (string, error) {
	if s.app == nil {
		return "", errors.New("file picker unavailable in server mode")
	}
	result, err := s.app.Dialog.SaveFile().
		SetMessage("Export identity").
		SetFilename("transport_identity").
		AddFilter("Reticulum identity", "transport_identity").
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *BrowserService) identityRegistry() (*rns.IdentityRegistry, error) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return nil, fmt.Errorf("reticulum not initialized")
	}
	reg := stack.Identities()
	if reg == nil {
		return nil, errors.New("identity registry not initialized")
	}
	return reg, nil
}

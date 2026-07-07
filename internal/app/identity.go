// SPDX-License-Identifier: MIT
package app

import (
	"errors"
	"fmt"
	"strings"

	"renbrowser/internal/rns"
)

var ErrIdentityPickerCanceled = errors.New("identity file selection canceled")

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
		return rns.IdentityRecord{}, identityOpError("create identity", err)
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
		return rns.IdentityRecord{}, identityOpError("import identity", err)
	}
	s.log("info", "identity imported", record.Hash)
	return record, nil
}

func (s *BrowserService) ExportIdentity(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return rns.ErrIdentityIDInvalid
	}
	reg, err := s.identityRegistry()
	if err != nil {
		return err
	}
	dest, err := s.PickIdentityExportPath()
	if err != nil {
		return err
	}
	if err := reg.Export(id, dest); err != nil {
		return identityOpError("export identity", err)
	}
	s.log("info", "identity exported", dest)
	return nil
}

func (s *BrowserService) SetActiveIdentity(id string) (rns.IdentityRecord, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return rns.IdentityRecord{}, rns.ErrIdentityIDInvalid
	}
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return rns.IdentityRecord{}, fmt.Errorf("reticulum not initialized")
	}
	if err := stack.SwitchIdentity(id); err != nil {
		return rns.IdentityRecord{}, identityOpError("switch identity", err)
	}
	reg := stack.Identities()
	if reg == nil {
		return rns.IdentityRecord{}, errors.New("identity registry not initialized")
	}
	record, err := reg.ActiveRecord()
	if err != nil {
		return rns.IdentityRecord{}, identityOpError("read active identity", err)
	}
	s.log("info", "active identity switched", record.Hash)
	if s.app != nil {
		s.app.Event.Emit("identity:changed", record)
	}
	return record, nil
}

func (s *BrowserService) RenameIdentity(id, name string) (rns.IdentityRecord, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return rns.IdentityRecord{}, rns.ErrIdentityIDInvalid
	}
	reg, err := s.identityRegistry()
	if err != nil {
		return rns.IdentityRecord{}, err
	}
	record, err := reg.Rename(id, name)
	if err != nil {
		return rns.IdentityRecord{}, identityOpError("rename identity", err)
	}
	return record, nil
}

func (s *BrowserService) DeleteIdentity(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return rns.ErrIdentityIDInvalid
	}
	reg, err := s.identityRegistry()
	if err != nil {
		return err
	}
	if err := reg.Delete(id); err != nil {
		return identityOpError("delete identity", err)
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
	if strings.TrimSpace(result) == "" {
		return "", ErrIdentityPickerCanceled
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
	if strings.TrimSpace(result) == "" {
		return "", ErrIdentityPickerCanceled
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

func identityOpError(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, rns.ErrIdentityNotFound) ||
		errors.Is(err, rns.ErrIdentityNameEmpty) ||
		errors.Is(err, rns.ErrIdentityNameTooLong) ||
		errors.Is(err, rns.ErrIdentityIDInvalid) ||
		errors.Is(err, rns.ErrIdentityDuplicate) ||
		errors.Is(err, rns.ErrCannotDeleteActive) ||
		errors.Is(err, rns.ErrCannotDeleteLast) ||
		errors.Is(err, rns.ErrIdentityAlreadyActive) ||
		errors.Is(err, rns.ErrRegistryCorrupt) ||
		errors.Is(err, rns.ErrInvalidIdentityFile) ||
		errors.Is(err, rns.ErrStorageDirEmpty) {
		return err
	}
	return fmt.Errorf("%s: %w", op, err)
}

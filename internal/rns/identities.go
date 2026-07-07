// SPDX-License-Identifier: MIT
package rns

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"quad4/reticulum-go/pkg/identity"
)

const identityRegistryVersion = 1

var (
	ErrIdentityNotFound      = errors.New("identity not found")
	ErrIdentityNameEmpty     = errors.New("identity name is required")
	ErrCannotDeleteActive    = errors.New("cannot delete the active identity")
	ErrCannotDeleteLast      = errors.New("cannot delete the only identity")
	ErrIdentityAlreadyActive = errors.New("identity is already active")
)

// IdentityRecord describes a stored transport identity for the settings UI.
type IdentityRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	CreatedAt int64  `json:"createdAt"`
	Active    bool   `json:"active"`
}

type identityRegistryFile struct {
	Version  int                    `json:"version"`
	ActiveID string                 `json:"activeId"`
	Items    []identityRegistryItem `json:"items"`
}

type identityRegistryItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	CreatedAt int64  `json:"createdAt"`
}

// IdentityRegistry manages multiple transport identities under the Reticulum storage dir.
type IdentityRegistry struct {
	mu         sync.Mutex
	storageDir string
	data       identityRegistryFile
}

func identitiesDir(storageDir string) string {
	return filepath.Join(storageDir, "identities")
}

func identityKeyPath(storageDir, id string) string {
	return filepath.Join(identitiesDir(storageDir), id)
}

func identityRegistryPath(storageDir string) string {
	return filepath.Join(storageDir, "identities.json")
}

func transportIdentityPath(storageDir string) string {
	return filepath.Join(storageDir, "transport_identity")
}

// OpenIdentityRegistry loads or creates the identity registry for a storage directory.
func OpenIdentityRegistry(storageDir string) (*IdentityRegistry, error) {
	if err := os.MkdirAll(storageDir, 0o700); err != nil {
		return nil, fmt.Errorf("storage dir: %w", err)
	}
	if err := os.MkdirAll(identitiesDir(storageDir), 0o700); err != nil {
		return nil, fmt.Errorf("identities dir: %w", err)
	}

	reg := &IdentityRegistry{storageDir: storageDir}
	path := identityRegistryPath(storageDir)
	raw, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := reg.migrateLegacyTransportIdentity(); err != nil {
			return nil, err
		}
		if len(reg.data.Items) == 0 {
			if _, err := reg.Create("Default"); err != nil {
				return nil, err
			}
		}
		return reg, reg.persist()
	}

	if err := json.Unmarshal(raw, &reg.data); err != nil {
		return nil, fmt.Errorf("parse identities registry: %w", err)
	}
	if reg.data.Version == 0 {
		reg.data.Version = identityRegistryVersion
	}
	if err := reg.reconcileActiveKeyFile(); err != nil {
		return nil, err
	}
	return reg, nil
}

func (r *IdentityRegistry) migrateLegacyTransportIdentity() error {
	legacyPath := transportIdentityPath(r.storageDir)
	ident, err := identity.FromFile(legacyPath)
	if err != nil {
		return nil
	}
	id, err := newIdentityID()
	if err != nil {
		return err
	}
	keyPath := identityKeyPath(r.storageDir, id)
	if err := ident.ToFile(keyPath); err != nil {
		return fmt.Errorf("migrate legacy identity: %w", err)
	}
	r.data = identityRegistryFile{
		Version:  identityRegistryVersion,
		ActiveID: id,
		Items: []identityRegistryItem{{
			ID:        id,
			Name:      "Default",
			Hash:      ident.GetHexHash(),
			CreatedAt: time.Now().Unix(),
		}},
	}
	return r.syncTransportIdentityFile(ident)
}

func (r *IdentityRegistry) reconcileActiveKeyFile() error {
	if r.data.ActiveID == "" {
		if len(r.data.Items) == 0 {
			return nil
		}
		r.data.ActiveID = r.data.Items[0].ID
	}
	item, ok := r.findItem(r.data.ActiveID)
	if !ok {
		return fmt.Errorf("active identity %q not found in registry", r.data.ActiveID)
	}
	keyPath := identityKeyPath(r.storageDir, item.ID)
	ident, err := identity.FromFile(keyPath)
	if err != nil {
		return fmt.Errorf("load active identity %q: %w", item.ID, err)
	}
	if item.Hash != ident.GetHexHash() {
		item.Hash = ident.GetHexHash()
	}
	return r.syncTransportIdentityFile(ident)
}

func (r *IdentityRegistry) List() []IdentityRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]IdentityRecord, 0, len(r.data.Items))
	for _, item := range r.data.Items {
		out = append(out, r.recordFromItem(item))
	}
	return out
}

func (r *IdentityRegistry) ActiveRecord() (IdentityRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.findItem(r.data.ActiveID)
	if !ok {
		return IdentityRecord{}, ErrIdentityNotFound
	}
	return r.recordFromItem(item), nil
}

func (r *IdentityRegistry) LoadActive() (*identity.Identity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.findItem(r.data.ActiveID)
	if !ok {
		return nil, ErrIdentityNotFound
	}
	return identity.FromFile(identityKeyPath(r.storageDir, item.ID))
}

func (r *IdentityRegistry) Create(name string) (IdentityRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return IdentityRecord{}, ErrIdentityNameEmpty
	}
	ident, err := identity.New()
	if err != nil {
		return IdentityRecord{}, err
	}
	id, err := newIdentityID()
	if err != nil {
		return IdentityRecord{}, err
	}
	keyPath := identityKeyPath(r.storageDir, id)
	if err := ident.ToFile(keyPath); err != nil {
		return IdentityRecord{}, err
	}
	item := identityRegistryItem{
		ID:        id,
		Name:      name,
		Hash:      ident.GetHexHash(),
		CreatedAt: time.Now().Unix(),
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.Items = append(r.data.Items, item)
	if r.data.ActiveID == "" {
		r.data.ActiveID = id
		if err := r.syncTransportIdentityFile(ident); err != nil {
			return IdentityRecord{}, err
		}
	}
	if err := r.persistLocked(); err != nil {
		return IdentityRecord{}, err
	}
	return r.recordFromItem(item), nil
}

func (r *IdentityRegistry) ImportFromFile(srcPath, name string) (IdentityRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return IdentityRecord{}, ErrIdentityNameEmpty
	}
	ident, err := identity.FromFile(srcPath)
	if err != nil {
		return IdentityRecord{}, err
	}
	id, err := newIdentityID()
	if err != nil {
		return IdentityRecord{}, err
	}
	keyPath := identityKeyPath(r.storageDir, id)
	if err := ident.ToFile(keyPath); err != nil {
		return IdentityRecord{}, err
	}
	item := identityRegistryItem{
		ID:        id,
		Name:      name,
		Hash:      ident.GetHexHash(),
		CreatedAt: time.Now().Unix(),
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.Items = append(r.data.Items, item)
	if r.data.ActiveID == "" {
		r.data.ActiveID = id
		if err := r.syncTransportIdentityFile(ident); err != nil {
			return IdentityRecord{}, err
		}
	}
	if err := r.persistLocked(); err != nil {
		return IdentityRecord{}, err
	}
	return r.recordFromItem(item), nil
}

func (r *IdentityRegistry) Export(id, destPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.findItem(id)
	if !ok {
		return ErrIdentityNotFound
	}
	src := identityKeyPath(r.storageDir, item.ID)
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if len(data) != 64 {
		return fmt.Errorf("invalid identity key file")
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o700); err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0o600)
}

func (r *IdentityRegistry) SetActive(id string) (*identity.Identity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.findItem(id)
	if !ok {
		return nil, ErrIdentityNotFound
	}
	if r.data.ActiveID == id {
		return nil, ErrIdentityAlreadyActive
	}
	ident, err := identity.FromFile(identityKeyPath(r.storageDir, item.ID))
	if err != nil {
		return nil, err
	}
	r.data.ActiveID = id
	if err := r.syncTransportIdentityFile(ident); err != nil {
		return nil, err
	}
	if err := r.persistLocked(); err != nil {
		return nil, err
	}
	return ident, nil
}

func (r *IdentityRegistry) Rename(id, name string) (IdentityRecord, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return IdentityRecord{}, ErrIdentityNameEmpty
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.findItem(id)
	if !ok {
		return IdentityRecord{}, ErrIdentityNotFound
	}
	item.Name = name
	if err := r.persistLocked(); err != nil {
		return IdentityRecord{}, err
	}
	return r.recordFromItem(item), nil
}

func (r *IdentityRegistry) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.data.Items) <= 1 {
		return ErrCannotDeleteLast
	}
	if r.data.ActiveID == id {
		return ErrCannotDeleteActive
	}
	idx := -1
	for i, item := range r.data.Items {
		if item.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrIdentityNotFound
	}
	keyPath := identityKeyPath(r.storageDir, id)
	_ = os.Remove(keyPath)
	r.data.Items = append(r.data.Items[:idx], r.data.Items[idx+1:]...)
	return r.persistLocked()
}

func (r *IdentityRegistry) syncTransportIdentityFile(ident *identity.Identity) error {
	if ident == nil {
		return errors.New("nil identity")
	}
	return ident.ToFile(transportIdentityPath(r.storageDir))
}

func (r *IdentityRegistry) persist() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.persistLocked()
}

func (r *IdentityRegistry) persistLocked() error {
	if r.data.Version == 0 {
		r.data.Version = identityRegistryVersion
	}
	raw, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}
	path := identityRegistryPath(r.storageDir)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (r *IdentityRegistry) findItem(id string) (identityRegistryItem, bool) {
	for _, item := range r.data.Items {
		if item.ID == id {
			return item, true
		}
	}
	return identityRegistryItem{}, false
}

func (r *IdentityRegistry) recordFromItem(item identityRegistryItem) IdentityRecord {
	return IdentityRecord{
		ID:        item.ID,
		Name:      item.Name,
		Hash:      item.Hash,
		CreatedAt: item.CreatedAt,
		Active:    item.ID == r.data.ActiveID,
	}
}

func newIdentityID() (string, error) {
	var buf [16]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

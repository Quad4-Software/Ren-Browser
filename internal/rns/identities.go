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
	"unicode/utf8"

	"quad4/reticulum-go/pkg/identity"
)

const (
	identityRegistryVersion = 1
	maxIdentityNameLen      = 128
	identityIDLen           = 32
	identityKeySize         = 64
)

var (
	ErrIdentityNotFound      = errors.New("identity not found")
	ErrIdentityNameEmpty     = errors.New("identity name is required")
	ErrIdentityNameTooLong   = errors.New("identity name is too long")
	ErrIdentityIDInvalid     = errors.New("identity id is invalid")
	ErrIdentityDuplicate     = errors.New("identity already exists")
	ErrCannotDeleteActive    = errors.New("cannot delete the active identity")
	ErrCannotDeleteLast      = errors.New("cannot delete the only identity")
	ErrIdentityAlreadyActive = errors.New("identity is already active")
	ErrRegistryCorrupt       = errors.New("identity registry is corrupt")
	ErrInvalidIdentityFile   = errors.New("invalid identity key file")
	ErrStorageDirEmpty       = errors.New("identity storage directory is required")
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

func validateStorageDir(storageDir string) error {
	storageDir = strings.TrimSpace(storageDir)
	if storageDir == "" {
		return ErrStorageDirEmpty
	}
	return nil
}

func validateIdentityID(id string) error {
	if id == "" {
		return ErrIdentityIDInvalid
	}
	if len(id) != identityIDLen {
		return ErrIdentityIDInvalid
	}
	if strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return ErrIdentityIDInvalid
	}
	for _, c := range id {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return ErrIdentityIDInvalid
		}
	}
	return nil
}

func validateIdentityName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrIdentityNameEmpty
	}
	if utf8.RuneCountInString(name) > maxIdentityNameLen {
		return ErrIdentityNameTooLong
	}
	return nil
}

func validateIdentityKeyData(data []byte) error {
	if len(data) != identityKeySize {
		return fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidIdentityFile, identityKeySize, len(data))
	}
	return nil
}

func ensureIdentityKeyUnderDir(storageDir, id string) error {
	if err := validateIdentityID(id); err != nil {
		return err
	}
	base, err := filepath.Abs(identitiesDir(storageDir))
	if err != nil {
		return err
	}
	target, err := filepath.Abs(identityKeyPath(storageDir, id))
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(base, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ErrIdentityIDInvalid
	}
	return nil
}

// OpenIdentityRegistry loads or creates the identity registry for a storage directory.
func OpenIdentityRegistry(storageDir string) (*IdentityRegistry, error) {
	if err := validateStorageDir(storageDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(storageDir, 0o700); err != nil {
		return nil, fmt.Errorf("storage dir: %w", err)
	}
	if err := os.MkdirAll(identitiesDir(storageDir), 0o700); err != nil {
		return nil, fmt.Errorf("identities dir: %w", err)
	}

	reg := &IdentityRegistry{storageDir: storageDir}
	raw, err := readStorageFile(storageDir, "identities.json")
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("read identity registry: %w", err)
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

	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: empty registry file", ErrRegistryCorrupt)
	}
	if err := json.Unmarshal(raw, &reg.data); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRegistryCorrupt, err)
	}
	if err := reg.validateRegistryData(); err != nil {
		return nil, err
	}
	if reg.data.Version == 0 {
		reg.data.Version = identityRegistryVersion
	}
	if err := reg.reconcileActiveKeyFile(); err != nil {
		return nil, err
	}
	return reg, nil
}

func (r *IdentityRegistry) validateRegistryData() error {
	seenIDs := make(map[string]struct{}, len(r.data.Items))
	seenHashes := make(map[string]struct{}, len(r.data.Items))
	for _, item := range r.data.Items {
		if err := validateIdentityID(item.ID); err != nil {
			return fmt.Errorf("%w: invalid item id %q", ErrRegistryCorrupt, item.ID)
		}
		if err := validateIdentityName(item.Name); err != nil {
			return fmt.Errorf("%w: invalid item name for %q", ErrRegistryCorrupt, item.ID)
		}
		if item.Hash == "" {
			return fmt.Errorf("%w: missing hash for %q", ErrRegistryCorrupt, item.ID)
		}
		if _, ok := seenIDs[item.ID]; ok {
			return fmt.Errorf("%w: duplicate id %q", ErrRegistryCorrupt, item.ID)
		}
		seenIDs[item.ID] = struct{}{}
		if _, ok := seenHashes[item.Hash]; ok {
			return fmt.Errorf("%w: duplicate hash %q", ErrRegistryCorrupt, item.Hash)
		}
		seenHashes[item.Hash] = struct{}{}
	}
	if r.data.ActiveID != "" {
		if err := validateIdentityID(r.data.ActiveID); err != nil {
			return fmt.Errorf("%w: invalid active id %q", ErrRegistryCorrupt, r.data.ActiveID)
		}
		if _, ok := seenIDs[r.data.ActiveID]; !ok {
			return fmt.Errorf("%w: active id %q not found", ErrRegistryCorrupt, r.data.ActiveID)
		}
	}
	return nil
}

func (r *IdentityRegistry) migrateLegacyTransportIdentity() error {
	legacyData, err := readStorageFile(r.storageDir, "transport_identity")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	ident, err := identity.FromBytes(legacyData)
	if err != nil {
		return nil
	}
	id, err := newIdentityID()
	if err != nil {
		return err
	}
	if err := writeIdentityToStorage(r.storageDir, id, ident); err != nil {
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
			return fmt.Errorf("%w: registry has no identities", ErrRegistryCorrupt)
		}
		r.data.ActiveID = r.data.Items[0].ID
	}
	item, ok := r.findItem(r.data.ActiveID)
	if !ok {
		return fmt.Errorf("%w: active identity %q not found in registry", ErrRegistryCorrupt, r.data.ActiveID)
	}
	if err := ensureIdentityKeyUnderDir(r.storageDir, item.ID); err != nil {
		return err
	}
	ident, err := loadIdentityFromStorage(r.storageDir, item.ID)
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
	if err := ensureIdentityKeyUnderDir(r.storageDir, item.ID); err != nil {
		return nil, err
	}
	return loadIdentityFromStorage(r.storageDir, item.ID)
}

func (r *IdentityRegistry) Create(name string) (IdentityRecord, error) {
	if err := validateIdentityName(name); err != nil {
		return IdentityRecord{}, err
	}
	name = strings.TrimSpace(name)
	ident, err := identity.New()
	if err != nil {
		return IdentityRecord{}, fmt.Errorf("generate identity: %w", err)
	}
	id, err := newIdentityID()
	if err != nil {
		return IdentityRecord{}, err
	}
	if err := writeIdentityToStorage(r.storageDir, id, ident); err != nil {
		return IdentityRecord{}, fmt.Errorf("write identity key: %w", err)
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
			r.rollbackCreateLocked(id)
			return IdentityRecord{}, fmt.Errorf("sync transport identity: %w", err)
		}
	}
	if err := r.persistLocked(); err != nil {
		r.rollbackCreateLocked(id)
		return IdentityRecord{}, fmt.Errorf("persist registry: %w", err)
	}
	return r.recordFromItem(item), nil
}

func (r *IdentityRegistry) ImportFromFile(srcPath, name string) (IdentityRecord, error) {
	srcPath = strings.TrimSpace(srcPath)
	if srcPath == "" {
		return IdentityRecord{}, fmt.Errorf("%w: source path is required", ErrInvalidIdentityFile)
	}
	if err := validateIdentityName(name); err != nil {
		return IdentityRecord{}, err
	}
	name = strings.TrimSpace(name)
	ident, err := identity.FromFile(srcPath)
	if err != nil {
		return IdentityRecord{}, fmt.Errorf("read identity file: %w", err)
	}
	hash := ident.GetHexHash()
	r.mu.Lock()
	if r.hasHashLocked(hash) {
		r.mu.Unlock()
		return IdentityRecord{}, fmt.Errorf("%w: %s", ErrIdentityDuplicate, hash)
	}
	r.mu.Unlock()

	id, err := newIdentityID()
	if err != nil {
		return IdentityRecord{}, err
	}
	if err := writeIdentityToStorage(r.storageDir, id, ident); err != nil {
		return IdentityRecord{}, fmt.Errorf("write identity key: %w", err)
	}
	item := identityRegistryItem{
		ID:        id,
		Name:      name,
		Hash:      hash,
		CreatedAt: time.Now().Unix(),
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.hasHashLocked(hash) {
		_ = removeIdentityKeyStorage(r.storageDir, id)
		return IdentityRecord{}, fmt.Errorf("%w: %s", ErrIdentityDuplicate, hash)
	}
	r.data.Items = append(r.data.Items, item)
	if r.data.ActiveID == "" {
		r.data.ActiveID = id
		if err := r.syncTransportIdentityFile(ident); err != nil {
			r.rollbackCreateLocked(id)
			return IdentityRecord{}, fmt.Errorf("sync transport identity: %w", err)
		}
	}
	if err := r.persistLocked(); err != nil {
		r.rollbackCreateLocked(id)
		return IdentityRecord{}, fmt.Errorf("persist registry: %w", err)
	}
	return r.recordFromItem(item), nil
}

func (r *IdentityRegistry) Export(id, destPath string) error {
	destPath = strings.TrimSpace(destPath)
	if destPath == "" {
		return errors.New("export path is required")
	}
	if err := validateIdentityID(id); err != nil {
		return err
	}
	r.mu.Lock()
	item, ok := r.findItem(id)
	r.mu.Unlock()
	if !ok {
		return ErrIdentityNotFound
	}
	if err := ensureIdentityKeyUnderDir(r.storageDir, item.ID); err != nil {
		return err
	}
	data, err := readIdentityKeyBytes(r.storageDir, item.ID)
	if err != nil {
		return fmt.Errorf("read identity key: %w", err)
	}
	if err := atomicWriteExportFile(destPath, data); err != nil {
		return err
	}
	return nil
}

func (r *IdentityRegistry) SetActive(id string) (*identity.Identity, error) {
	if err := validateIdentityID(id); err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.findItem(id)
	if !ok {
		return nil, ErrIdentityNotFound
	}
	if r.data.ActiveID == id {
		return nil, ErrIdentityAlreadyActive
	}
	if err := ensureIdentityKeyUnderDir(r.storageDir, item.ID); err != nil {
		return nil, err
	}
	ident, err := loadIdentityFromStorage(r.storageDir, item.ID)
	if err != nil {
		return nil, fmt.Errorf("load identity %q: %w", id, err)
	}
	previousID := r.data.ActiveID
	var previousIdent *identity.Identity
	if previousID != "" {
		previousIdent, _ = loadIdentityFromStorage(r.storageDir, previousID)
	}
	r.data.ActiveID = id
	if err := r.syncTransportIdentityFile(ident); err != nil {
		r.data.ActiveID = previousID
		if previousIdent != nil {
			_ = r.syncTransportIdentityFile(previousIdent)
		}
		return nil, fmt.Errorf("sync transport identity: %w", err)
	}
	if err := r.persistLocked(); err != nil {
		r.data.ActiveID = previousID
		if previousIdent != nil {
			_ = r.syncTransportIdentityFile(previousIdent)
		}
		return nil, fmt.Errorf("persist registry: %w", err)
	}
	return ident, nil
}

func (r *IdentityRegistry) Rename(id, name string) (IdentityRecord, error) {
	if err := validateIdentityID(id); err != nil {
		return IdentityRecord{}, err
	}
	if err := validateIdentityName(name); err != nil {
		return IdentityRecord{}, err
	}
	name = strings.TrimSpace(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.findItem(id)
	if !ok {
		return IdentityRecord{}, ErrIdentityNotFound
	}
	item.Name = name
	if err := r.persistLocked(); err != nil {
		return IdentityRecord{}, fmt.Errorf("persist registry: %w", err)
	}
	return r.recordFromItem(item), nil
}

func (r *IdentityRegistry) Delete(id string) error {
	if err := validateIdentityID(id); err != nil {
		return err
	}
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
	if err := removeIdentityKeyStorage(r.storageDir, id); err != nil {
		return fmt.Errorf("remove identity key: %w", err)
	}
	r.data.Items = append(r.data.Items[:idx], r.data.Items[idx+1:]...)
	if err := r.persistLocked(); err != nil {
		return fmt.Errorf("persist registry: %w", err)
	}
	return nil
}

func (r *IdentityRegistry) rollbackCreateLocked(id string) {
	for i, item := range r.data.Items {
		if item.ID == id {
			r.data.Items = append(r.data.Items[:i], r.data.Items[i+1:]...)
			break
		}
	}
	if r.data.ActiveID == id {
		r.data.ActiveID = ""
		if len(r.data.Items) > 0 {
			r.data.ActiveID = r.data.Items[0].ID
		}
	}
	_ = removeIdentityKeyStorage(r.storageDir, id)
}

func (r *IdentityRegistry) hasHashLocked(hash string) bool {
	for _, item := range r.data.Items {
		if item.Hash == hash {
			return true
		}
	}
	return false
}

func (r *IdentityRegistry) syncTransportIdentityFile(ident *identity.Identity) error {
	if ident == nil {
		return errors.New("nil identity")
	}
	privateKey, err := ident.GetPrivateKey()
	if err != nil {
		return err
	}
	return atomicWriteStorageFile(r.storageDir, "transport_identity", privateKey, 0o600)
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
	return atomicWriteStorageFile(r.storageDir, "identities.json", raw, 0o600)
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

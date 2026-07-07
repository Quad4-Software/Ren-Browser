// SPDX-License-Identifier: MIT
package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

const trustedPublishersDigestKey = "plugins.trustedPublishersDigest"

var (
	trustedIntegrityMu      sync.RWMutex
	trustedIntegrityStore   SettingsStore
	userTrustedTampered     bool
	userTrustedTamperReason string
)

type SettingsStore interface {
	GetSetting(key string) (string, error)
	SetSetting(key string, value string) error
}

func InitTrustedPublishersIntegrity(store SettingsStore) error {
	trustedIntegrityMu.Lock()
	trustedIntegrityStore = store
	trustedIntegrityMu.Unlock()
	return VerifyUserTrustedPublishersIntegrity(store)
}

func UserTrustedPublishersTampered() (bool, string) {
	trustedIntegrityMu.RLock()
	defer trustedIntegrityMu.RUnlock()
	return userTrustedTampered, userTrustedTamperReason
}

func VerifyUserTrustedPublishersIntegrity(store SettingsStore) error {
	if store == nil {
		return nil
	}
	path := userTrustedPublishersPath()
	raw, err := os.ReadFile(path) // #nosec G304 -- path under user data dir
	if os.IsNotExist(err) {
		trustedIntegrityMu.Lock()
		userTrustedTampered = false
		userTrustedTamperReason = ""
		trustedIntegrityMu.Unlock()
		return nil
	}
	if err != nil {
		return err
	}

	currentDigest, err := digestUserTrustedPublishersRaw(raw)
	if err != nil {
		return err
	}
	storedDigest, err := store.GetSetting(trustedPublishersDigestKey)
	if err != nil {
		return err
	}
	storedDigest = strings.TrimSpace(storedDigest)

	trustedIntegrityMu.Lock()
	defer trustedIntegrityMu.Unlock()

	if storedDigest == "" {
		userTrustedTampered = false
		userTrustedTamperReason = ""
		return store.SetSetting(trustedPublishersDigestKey, currentDigest)
	}
	if storedDigest != currentDigest {
		userTrustedTampered = true
		userTrustedTamperReason = "trusted publisher list was modified outside RenBrowser"
		return nil
	}
	userTrustedTampered = false
	userTrustedTamperReason = ""
	return nil
}

func updateUserTrustedPublishersDigest(store SettingsStore) error {
	if store == nil {
		return nil
	}
	path := userTrustedPublishersPath()
	raw, err := os.ReadFile(path) // #nosec G304 -- path under user data dir
	if os.IsNotExist(err) {
		return store.SetSetting(trustedPublishersDigestKey, "")
	}
	if err != nil {
		return err
	}
	digest, err := digestUserTrustedPublishersRaw(raw)
	if err != nil {
		return err
	}
	if err := store.SetSetting(trustedPublishersDigestKey, digest); err != nil {
		return err
	}
	trustedIntegrityMu.Lock()
	userTrustedTampered = false
	userTrustedTamperReason = ""
	trustedIntegrityMu.Unlock()
	return nil
}

func digestUserTrustedPublishersRaw(raw []byte) (string, error) {
	var file trustedSignersFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return "", fmt.Errorf("invalid trusted publishers file: %w", err)
	}
	canonical, err := json.Marshal(file)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

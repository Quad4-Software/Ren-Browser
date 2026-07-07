// SPDX-License-Identifier: MIT
package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"renbrowser/internal/paths"
)

var userTrustedMu sync.RWMutex

func userTrustedPublishersPath() string {
	return paths.Join(".renbrowser", "trusted_publishers.json")
}

func loadUserTrustedPublishers() []TrustedPublisher {
	if tampered, _ := UserTrustedPublishersTampered(); tampered {
		return nil
	}
	userTrustedMu.RLock()
	defer userTrustedMu.RUnlock()
	raw, err := os.ReadFile(userTrustedPublishersPath()) // #nosec G304 -- path under user data dir
	if err != nil {
		return nil
	}
	var file trustedSignersFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil
	}
	return file.Publishers
}

func AddUserTrustedPublisher(identity, name string) error {
	return AddUserTrustedPublisherWithStore(trustedIntegrityStore, identity, name)
}

func AddUserTrustedPublisherWithStore(store SettingsStore, identity, name string) error {
	identity = strings.ToLower(strings.TrimSpace(identity))
	if identity == "" {
		return fmt.Errorf("publisher identity is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = identity
	}

	userTrustedMu.Lock()
	defer userTrustedMu.Unlock()

	path := userTrustedPublishersPath()
	var file trustedSignersFile
	if raw, err := os.ReadFile(path); err == nil { // #nosec G304 -- path under user data dir
		_ = json.Unmarshal(raw, &file)
	}
	for _, pub := range file.Publishers {
		if strings.ToLower(strings.TrimSpace(pub.Identity)) == identity {
			return nil
		}
	}
	file.Publishers = append(file.Publishers, TrustedPublisher{
		Identity: identity,
		Name:     name,
	})
	raw, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return err
	}
	return updateUserTrustedPublishersDigest(store)
}

func ListTrustedPublishers() []TrustedPublisher {
	seen := make(map[string]struct{})
	var out []TrustedPublisher
	appendPub := func(pub TrustedPublisher) {
		id := strings.ToLower(strings.TrimSpace(pub.Identity))
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, TrustedPublisher{
			Identity: id,
			Name:     strings.TrimSpace(pub.Name),
		})
	}

	var bundled trustedSignersFile
	if err := json.Unmarshal(trustedSignersJSON, &bundled); err == nil {
		for _, pub := range bundled.Publishers {
			appendPub(pub)
		}
	}
	for _, pub := range loadUserTrustedPublishers() {
		appendPub(pub)
	}
	return out
}

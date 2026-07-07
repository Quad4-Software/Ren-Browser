// SPDX-License-Identifier: MIT
package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const integrityTamperMessage = "extension files were modified outside RenBrowser"

func ComputeDirIntegrityHash(dir string) (string, error) {
	payload, err := canonicalDirPayload(dir)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func VerifyDirIntegrity(dir, expectedHash string) (bool, string, error) {
	if expectedHash == "" {
		return true, "", nil
	}
	current, err := ComputeDirIntegrityHash(dir)
	if err != nil {
		return false, "", err
	}
	if current != expectedHash {
		return false, current, nil
	}
	return true, current, nil
}

func IntegrityTamperError() error {
	return fmt.Errorf("%s", integrityTamperMessage)
}

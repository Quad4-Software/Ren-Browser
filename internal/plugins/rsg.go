// SPDX-License-Identifier: MIT
package plugins

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"

	"quad4/msgpack/v5/pkg/msgpack"
	"quad4/reticulum-go/pkg/cryptography"
	"quad4/reticulum-go/pkg/identity"
)

const (
	ed25519SigLen = ed25519.SignatureSize
)

var (
	errRSGTooShort  = errors.New("rsg data too short")
	errRSGEnvelope  = errors.New("invalid rsg envelope")
	errRSGHashType  = errors.New("unsupported rsg hash type")
	errRSGHash      = errors.New("rsg hash mismatch")
	errRSGSigner    = errors.New("rsg signer mismatch")
	errRSGSignature = errors.New("invalid rsg signature")
	errRSGPublicKey = errors.New("invalid rsg public key")
)

type rsgEnvelope struct {
	HashType string         `msgpack:"hashtype"`
	Hash     []byte         `msgpack:"hash"`
	Meta     map[string]any `msgpack:"meta"`
	Message  []byte         `msgpack:"message,omitempty"`
}

// ValidateRSG verifies a Reticulum Signature (.rsg) against message bytes.
// requiredSigner, when non-empty, must match the signer identity hash in the envelope.
func ValidateRSG(rsgData, message []byte, requiredSigner []byte) (signerHex string, err error) {
	if len(message) == 0 {
		return "", errors.New("no message for rsg validation")
	}
	if len(rsgData) < ed25519SigLen+1 {
		return "", errRSGTooShort
	}
	signature := rsgData[:ed25519SigLen]
	envelopeBytes := rsgData[ed25519SigLen:]

	var envelope rsgEnvelope
	if err := msgpack.Unmarshal(envelopeBytes, &envelope); err != nil {
		return "", fmt.Errorf("%w: %v", errRSGEnvelope, err)
	}
	if envelope.HashType != "sha256" {
		return "", errRSGHashType
	}
	if len(envelope.Hash) != 32 {
		return "", errRSGHash
	}
	messageHash := cryptography.Hash(message)
	if !bytes.Equal(envelope.Hash, messageHash) {
		return "", errRSGHash
	}
	if envelope.Meta == nil {
		return "", errRSGEnvelope
	}
	signerRaw, ok := metaBytes(envelope.Meta, "signer")
	if !ok || len(signerRaw) != identity.TruncatedHashLength/8 {
		return "", errRSGSigner
	}
	pubKeyRaw, ok := metaBytes(envelope.Meta, "pubkey")
	if !ok || len(pubKeyRaw) != identity.KeySize/8 {
		return "", errRSGPublicKey
	}
	signingIdentity := identity.FromPublicKey(pubKeyRaw)
	if signingIdentity == nil {
		return "", errRSGPublicKey
	}
	if !bytes.Equal(signingIdentity.Hash(), signerRaw) {
		return "", errRSGSigner
	}
	if len(requiredSigner) > 0 && !bytes.Equal(signingIdentity.Hash(), requiredSigner) {
		return "", errRSGSigner
	}
	if !signingIdentity.Verify(envelopeBytes, signature) {
		return "", errRSGSignature
	}
	return hex.EncodeToString(signingIdentity.Hash()), nil
}

func metaBytes(meta map[string]any, key string) ([]byte, bool) {
	raw, ok := meta[key]
	if !ok {
		return nil, false
	}
	switch v := raw.(type) {
	case []byte:
		return v, true
	case string:
		return []byte(v), true
	default:
		return nil, false
	}
}

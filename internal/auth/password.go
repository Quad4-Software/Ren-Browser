// SPDX-License-Identifier: MIT
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

var ErrInvalidPassword = errors.New("invalid password")

func HashPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", errors.New("password must not be empty")
	}
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads, b64Salt, b64Hash), nil
}

func VerifyPassword(encoded, password string) error {
	salt, hash, params, err := decodeArgon2ID(encoded)
	if err != nil {
		return ErrInvalidPassword
	}
	hashLen := len(hash)
	if hashLen == 0 || hashLen > math.MaxUint32 {
		return ErrInvalidPassword
	}
	other := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, uint32(hashLen))
	if subtle.ConstantTimeCompare(hash, other) != 1 {
		return ErrInvalidPassword
	}
	return nil
}

type argonParams struct {
	memory  uint32
	time    uint32
	threads uint8
}

func decodeArgon2ID(encoded string) (salt, hash []byte, params argonParams, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, params, errors.New("invalid argon2id hash")
	}
	var version int
	if _, err = fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, params, err
	}
	if version != argon2.Version {
		return nil, nil, params, errors.New("unsupported argon2 version")
	}
	var memory, time uint32
	var threads int
	if _, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return nil, nil, params, err
	}
	if threads < 1 || threads > 255 {
		return nil, nil, params, errors.New("invalid thread count")
	}
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, params, err
	}
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, params, err
	}
	params = argonParams{memory: memory, time: time, threads: uint8(threads)}
	return salt, hash, params, nil
}

func NewSessionToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func ParsePositiveInt(raw string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

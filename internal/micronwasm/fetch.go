// SPDX-License-Identifier: MIT
package micronwasm

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"renbrowser/internal/constants"
)

var shasumLine = regexp.MustCompile(`^([a-fA-F0-9]{64})\s+\*?(\S+)\s*$`)

type FetchResult struct {
	ReleaseTag string `json:"releaseTag"`
	WasmBase64 string `json:"wasmBase64"`
	Sha256Hex  string `json:"sha256Hex"`
}

func ParseShasums256ForFilename(text, filename string) (string, error) {
	if strings.TrimSpace(text) == "" || strings.TrimSpace(filename) == "" {
		return "", fmt.Errorf("empty shasums input")
	}
	for raw := range strings.SplitSeq(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := shasumLine.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		name := strings.TrimSpace(m[2])
		if name == filename || strings.HasSuffix(name, "/"+filename) {
			return strings.ToLower(m[1]), nil
		}
	}
	return "", fmt.Errorf("%s not listed in %s", filename, constants.MicronParserGoShasumsFilename)
}

func FetchVerifiedRelease(tag string) (FetchResult, error) {
	trimmed := strings.TrimSpace(tag)
	if trimmed == "" {
		return FetchResult{}, fmt.Errorf("release tag is required")
	}

	client := &http.Client{Timeout: 120 * time.Second}
	sumsURL := fmt.Sprintf("%s/%s/%s", constants.MicronParserGoReleaseDownloadBase, trimmed, constants.MicronParserGoShasumsFilename)
	sumsRes, err := client.Get(sumsURL)
	if err != nil {
		return FetchResult{}, fmt.Errorf("fetch shasums: %w", err)
	}
	defer sumsRes.Body.Close()
	if sumsRes.StatusCode < 200 || sumsRes.StatusCode >= 300 {
		return FetchResult{}, fmt.Errorf("fetch shasums: HTTP %d", sumsRes.StatusCode)
	}
	sumsBody, err := io.ReadAll(io.LimitReader(sumsRes.Body, 1<<20))
	if err != nil {
		return FetchResult{}, fmt.Errorf("read shasums: %w", err)
	}
	expectedHex, err := ParseShasums256ForFilename(string(sumsBody), constants.MicronParserGoWasmFilename)
	if err != nil {
		return FetchResult{}, err
	}

	wasmURL := fmt.Sprintf("%s/%s/%s", constants.MicronParserGoReleaseDownloadBase, trimmed, constants.MicronParserGoWasmFilename)
	wasmRes, err := client.Get(wasmURL)
	if err != nil {
		return FetchResult{}, fmt.Errorf("fetch wasm: %w", err)
	}
	defer wasmRes.Body.Close()
	if wasmRes.StatusCode < 200 || wasmRes.StatusCode >= 300 {
		return FetchResult{}, fmt.Errorf("fetch wasm: HTTP %d", wasmRes.StatusCode)
	}
	wasmBody, err := io.ReadAll(io.LimitReader(wasmRes.Body, constants.MicronParserGoMaxWasmBytes+1))
	if err != nil {
		return FetchResult{}, fmt.Errorf("read wasm: %w", err)
	}
	if len(wasmBody) > constants.MicronParserGoMaxWasmBytes {
		return FetchResult{}, fmt.Errorf("wasm exceeds maximum size (%d bytes)", constants.MicronParserGoMaxWasmBytes)
	}
	if len(wasmBody) < 4096 {
		return FetchResult{}, fmt.Errorf("wasm file is too small")
	}

	sum := sha256.Sum256(wasmBody)
	actualHex := hex.EncodeToString(sum[:])
	if actualHex != expectedHex {
		return FetchResult{}, fmt.Errorf("sha-256 mismatch after download")
	}

	return FetchResult{
		ReleaseTag: trimmed,
		WasmBase64: base64.StdEncoding.EncodeToString(wasmBody),
		Sha256Hex:  actualHex,
	}, nil
}

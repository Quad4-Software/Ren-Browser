// SPDX-License-Identifier: MIT
package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

const (
	wasmMagic           = "\x00asm"
	wasmVersion         = "\x01\x00\x00\x00"
	wasmSectionCustom   = byte(0)
	wasmSectionPlugin   = "renbrowser.plugin"
	wasmSectionFiles    = "renbrowser.files"
	maxWasmBundleBytes  = 32 * 1024 * 1024
	maxWasmEmbeddedFile = 8 * 1024 * 1024
	maxWasmEmbeddedKeys = 64
)

type WasmBundle struct {
	Manifest   Manifest
	Files      map[string]string
	WasmBinary []byte
	Signature  []byte
}

func ParseWasmBundle(data []byte) (WasmBundle, error) {
	if len(data) < 8 || string(data[:4]) != wasmMagic {
		return WasmBundle{}, errors.New("invalid wasm module")
	}
	if string(data[4:8]) != wasmVersion {
		return WasmBundle{}, errors.New("unsupported wasm version")
	}
	bundle := WasmBundle{WasmBinary: data, Files: map[string]string{}}
	offset := 8
	for offset < len(data) {
		sectionID := data[offset]
		offset++
		size, n, err := readULEB32(data[offset:])
		if err != nil {
			return WasmBundle{}, fmt.Errorf("read wasm section: %w", err)
		}
		offset += n
		if int(size) > len(data)-offset {
			return WasmBundle{}, errors.New("truncated wasm section")
		}
		payload := data[offset : offset+int(size)]
		offset += int(size)
		if sectionID != wasmSectionCustom {
			continue
		}
		name, rest, err := readByteVector(payload)
		if err != nil {
			continue
		}
		content, _, err := readByteVector(rest)
		if err != nil {
			continue
		}
		switch string(name) {
		case wasmSectionPlugin:
			if err := json.Unmarshal(content, &bundle.Manifest); err != nil {
				return WasmBundle{}, fmt.Errorf("parse embedded manifest: %w", err)
			}
		case wasmSectionFiles:
			var files map[string]string
			if err := json.Unmarshal(content, &files); err != nil {
				return WasmBundle{}, fmt.Errorf("parse embedded files: %w", err)
			}
			maps.Copy(bundle.Files, files)
		case wasmSectionSignature:
			bundle.Signature = append([]byte(nil), content...)
		}
	}
	return bundle, nil
}

func (b WasmBundle) ValidateEmbedded() error {
	if strings.TrimSpace(b.Manifest.ID) == "" {
		return fmt.Errorf("wasm module missing %s metadata", wasmSectionPlugin)
	}
	if err := ValidateManifest(b.Manifest); err != nil {
		return err
	}
	if strings.TrimSpace(b.Manifest.Backend) == "" {
		return fmt.Errorf("embedded manifest must set backend")
	}
	if len(b.Files) > maxWasmEmbeddedKeys {
		return fmt.Errorf("too many embedded files")
	}
	for name, content := range b.Files {
		if err := validateEmbeddedPath(name); err != nil {
			return err
		}
		if len(content) > maxWasmEmbeddedFile {
			return fmt.Errorf("embedded file %q too large", name)
		}
	}
	if b.Manifest.Main != "" {
		if _, ok := b.Files[b.Manifest.Main]; !ok {
			return fmt.Errorf("embedded files missing main entry %q", b.Manifest.Main)
		}
	}
	for _, panel := range b.Manifest.Contributes.Panels {
		if panel.Entry == "" {
			continue
		}
		if panel.Entry == b.Manifest.Main {
			continue
		}
		if _, ok := b.Files[panel.Entry]; !ok {
			return fmt.Errorf("embedded files missing panel entry %q", panel.Entry)
		}
	}
	return nil
}

func BundleWasm(wasm []byte, manifest Manifest, files map[string]string) ([]byte, error) {
	if len(wasm) < 8 || string(wasm[:4]) != wasmMagic {
		return nil, errors.New("invalid wasm module")
	}
	if len(wasm) > maxWasmBundleBytes {
		return nil, errors.New("wasm module too large")
	}
	if err := ValidateManifest(manifest); err != nil {
		return nil, err
	}
	if strings.TrimSpace(manifest.Backend) == "" {
		return nil, errors.New("manifest backend is required for wasm bundle")
	}
	if len(files) > maxWasmEmbeddedKeys {
		return nil, errors.New("too many embedded files")
	}
	out := append([]byte(nil), wasm...)
	manifestRaw, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	out, err = appendCustomSection(out, wasmSectionPlugin, manifestRaw)
	if err != nil {
		return nil, err
	}
	if len(files) > 0 {
		for name := range files {
			if err := validateEmbeddedPath(name); err != nil {
				return nil, err
			}
			if len(files[name]) > maxWasmEmbeddedFile {
				return nil, fmt.Errorf("embedded file %q too large", name)
			}
		}
		filesRaw, err := json.Marshal(files)
		if err != nil {
			return nil, err
		}
		out, err = appendCustomSection(out, wasmSectionFiles, filesRaw)
		if err != nil {
			return nil, err
		}
	}
	if len(out) > maxWasmBundleBytes {
		return nil, errors.New("bundled wasm module too large")
	}
	return out, nil
}

func appendCustomSection(wasm []byte, name string, data []byte) ([]byte, error) {
	if len(name) == 0 {
		return nil, errors.New("custom section name is required")
	}
	payload := append(encodeByteVector([]byte(name)), encodeByteVector(data)...)
	section := []byte{wasmSectionCustom}
	section = append(section, encodeULEB32(uint32(len(payload)))...) // #nosec G115 -- payload bounded by maxWasmBundleBytes
	section = append(section, payload...)
	return append(wasm, section...), nil
}

func validateEmbeddedPath(name string) error {
	clean := filepath.Clean(strings.ReplaceAll(name, `\`, "/"))
	if clean == "" || clean == "." || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return fmt.Errorf("invalid embedded path %q", name)
	}
	return nil
}

func writeWasmBundle(dest string, bundle WasmBundle) error {
	if err := os.MkdirAll(dest, 0o750); err != nil {
		return err
	}
	manifestPath := filepath.Join(dest, ManifestFileName)
	manifestRaw, err := json.MarshalIndent(bundle.Manifest, "", "  ")
	if err != nil {
		return err
	}
	manifestRaw = append(manifestRaw, '\n')
	if err := os.WriteFile(manifestPath, manifestRaw, 0o600); err != nil {
		return err
	}
	backendPath, err := safeZipJoin(dest, bundle.Manifest.Backend)
	if err != nil {
		return err
	}
	if err := os.WriteFile(backendPath, bundle.WasmBinary, 0o600); err != nil {
		return err
	}
	for name, content := range bundle.Files {
		if name == bundle.Manifest.Backend {
			continue
		}
		filePath, err := safeZipJoin(dest, name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(filePath), 0o750); err != nil {
			return err
		}
		if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
			return err
		}
	}
	return nil
}

func readULEB32(data []byte) (uint32, int, error) {
	var result uint32
	var shift uint
	for i, b := range data {
		if shift >= 35 {
			return 0, 0, errors.New("invalid uleb32")
		}
		result |= uint32(b&0x7f) << shift
		if b&0x80 == 0 {
			return result, i + 1, nil
		}
		shift += 7
	}
	return 0, 0, errors.New("truncated uleb32")
}

func encodeULEB32(v uint32) []byte {
	var out []byte
	for {
		b := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			b |= 0x80
		}
		out = append(out, b)
		if v == 0 {
			break
		}
	}
	return out
}

func readByteVector(data []byte) ([]byte, []byte, error) {
	size, n, err := readULEB32(data)
	if err != nil {
		return nil, nil, err
	}
	if int(size) > len(data)-n {
		return nil, nil, errors.New("truncated byte vector")
	}
	end := n + int(size)
	return data[n:end], data[end:], nil
}

func encodeByteVector(data []byte) []byte {
	out := encodeULEB32(uint32(len(data))) // #nosec G115 -- length capped by maxWasmBundleBytes
	return append(out, data...)
}

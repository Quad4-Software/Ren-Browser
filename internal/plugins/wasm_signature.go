// SPDX-License-Identifier: MIT
package plugins

import (
	"errors"
	"fmt"
)

// WasmPayloadWithoutSignature returns wasm module bytes excluding the signature custom section.
func WasmPayloadWithoutSignature(data []byte) ([]byte, error) {
	if len(data) < 8 || string(data[:4]) != wasmMagic {
		return nil, errors.New("invalid wasm module")
	}
	if string(data[4:8]) != wasmVersion {
		return nil, errors.New("unsupported wasm version")
	}
	out := append([]byte(nil), data[:8]...)
	offset := 8
	for offset < len(data) {
		sectionStart := offset
		sectionID := data[offset]
		offset++
		size, n, err := readULEB32(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("read wasm section: %w", err)
		}
		offset += n
		if int(size) > len(data)-offset {
			return nil, errors.New("truncated wasm section")
		}
		payload := data[offset : offset+int(size)]
		offset += int(size)
		if sectionID != wasmSectionCustom {
			out = append(out, data[sectionStart:offset]...)
			continue
		}
		name, _, err := readByteVector(payload)
		if err != nil {
			out = append(out, data[sectionStart:offset]...)
			continue
		}
		if string(name) == wasmSectionSignature {
			continue
		}
		out = append(out, data[sectionStart:offset]...)
	}
	return out, nil
}

// AppendWasmSignature embeds an RSG signature into a wasm bundle module.
func AppendWasmSignature(wasm []byte, signature []byte) ([]byte, error) {
	if len(signature) == 0 {
		return nil, errors.New("signature is required")
	}
	payload, err := WasmPayloadWithoutSignature(wasm)
	if err != nil {
		return nil, err
	}
	return appendCustomSection(payload, wasmSectionSignature, signature)
}

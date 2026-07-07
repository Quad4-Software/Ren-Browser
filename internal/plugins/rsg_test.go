// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRSGFromRnid(t *testing.T) {
	rsgPath := os.Getenv("REN_TEST_RSG")
	messagePath := os.Getenv("REN_TEST_RSG_MESSAGE")
	if rsgPath == "" || messagePath == "" {
		t.Skip("set REN_TEST_RSG and REN_TEST_RSG_MESSAGE to run interoperability test")
	}
	rsgData, err := os.ReadFile(rsgPath)
	if err != nil {
		t.Fatal(err)
	}
	message, err := os.ReadFile(messagePath)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ValidateRSG(rsgData, message, nil)
	if err != nil {
		t.Fatalf("ValidateRSG: %v", err)
	}
	if signer == "" {
		t.Fatal("expected signer hash")
	}
}

func TestDirSignatureRoundTrip(t *testing.T) {
	dir := t.TempDir()
	manifest := `{
  "manifestVersion": 1,
  "id": "test.signed",
  "name": "Signed",
  "version": "1.0.0",
  "contributes": {}
}`
	if err := os.WriteFile(filepath.Join(dir, ManifestFileName), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.js"), []byte("export {}"), 0o600); err != nil {
		t.Fatal(err)
	}

	payload, err := CanonicalDirPayload(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(payload) == 0 {
		t.Fatal("expected canonical payload")
	}

	unsigned := SignatureInfo{}
	if unsigned.Present {
		t.Fatal("expected unsigned plugin")
	}

	badRSG := make([]byte, 80)
	if err := WriteDirSignature(dir, badRSG); err != nil {
		t.Fatal(err)
	}
	invalid := VerifyDirSignature(dir)
	if !invalid.Present || invalid.Valid {
		t.Fatalf("invalid signature info = %#v", invalid)
	}
	if err := RequireValidSignature(invalid); err == nil {
		t.Fatal("expected invalid signature to be rejected")
	}
}

func TestWasmPayloadWithoutSignatureSection(t *testing.T) {
	base := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00,
	}
	wasm, err := appendCustomSection(base, wasmSectionSignature, []byte("sig"))
	if err != nil {
		t.Fatal(err)
	}
	stripped, err := WasmPayloadWithoutSignature(wasm)
	if err != nil {
		t.Fatal(err)
	}
	if len(stripped) != len(base) {
		t.Fatalf("stripped len=%d want %d", len(stripped), len(base))
	}
}

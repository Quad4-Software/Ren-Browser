// SPDX-License-Identifier: MIT
package plugins

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	SignatureFileName    = "renbrowser.plugin.rsg"
	wasmSectionSignature = "renbrowser.signature"
)

var fixedZipModTime = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

type trustedSignersFile struct {
	Publishers []TrustedPublisher `json:"publishers"`
}

type TrustedPublisher struct {
	Identity string `json:"identity"`
	Name     string `json:"name"`
}

type SignatureInfo struct {
	Present    bool   `json:"present"`
	Valid      bool   `json:"valid"`
	Signer     string `json:"signer,omitempty"`
	SignerName string `json:"signerName,omitempty"`
	Trusted    bool   `json:"trusted"`
	Error      string `json:"error,omitempty"`
}

func VerifyDirSignature(dir string) SignatureInfo {
	rsgPath := filepath.Join(dir, SignatureFileName)
	rsgData, err := os.ReadFile(rsgPath) // #nosec G304 -- plugin dir from user data
	if err != nil {
		if os.IsNotExist(err) {
			return SignatureInfo{}
		}
		return SignatureInfo{Present: true, Error: err.Error()}
	}
	payload, err := canonicalDirPayload(dir)
	if err != nil {
		return SignatureInfo{Present: true, Error: err.Error()}
	}
	return verifyRSGPayload(rsgData, payload)
}

func VerifyZipSignature(zipPath string) SignatureInfo {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return SignatureInfo{Error: err.Error()}
	}
	defer reader.Close()

	var rsgData []byte
	for _, f := range reader.File {
		clean := filepath.Clean(strings.ReplaceAll(f.Name, `\`, "/"))
		if filepath.Base(clean) == SignatureFileName {
			rc, openErr := f.Open()
			if openErr != nil {
				return SignatureInfo{Present: true, Error: openErr.Error()}
			}
			rsgData, err = io.ReadAll(rc)
			_ = rc.Close() // #nosec G104 -- read complete or error handled below
			if err != nil {
				return SignatureInfo{Present: true, Error: err.Error()}
			}
			break
		}
	}
	if len(rsgData) == 0 {
		return SignatureInfo{}
	}
	payload, err := canonicalZipPayload(reader.File)
	if err != nil {
		return SignatureInfo{Present: true, Error: err.Error()}
	}
	return verifyRSGPayload(rsgData, payload)
}

func VerifyWasmSignature(data []byte) SignatureInfo {
	bundle, err := ParseWasmBundle(data)
	if err != nil {
		return SignatureInfo{Error: err.Error()}
	}
	if len(bundle.Signature) == 0 {
		return SignatureInfo{}
	}
	payload, err := WasmPayloadWithoutSignature(data)
	if err != nil {
		return SignatureInfo{Present: true, Error: err.Error()}
	}
	return verifyRSGPayload(bundle.Signature, payload)
}

func verifyRSGPayload(rsgData, payload []byte) SignatureInfo {
	info := SignatureInfo{Present: true}
	signerHex, err := ValidateRSG(rsgData, payload, nil)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	info.Valid = true
	info.Signer = signerHex
	info.SignerName, info.Trusted = lookupTrustedPublisher(signerHex)
	return info
}

func lookupTrustedPublisher(signerHex string) (string, bool) {
	needle := strings.ToLower(strings.TrimSpace(signerHex))
	for _, pub := range ListTrustedPublishers() {
		if strings.ToLower(strings.TrimSpace(pub.Identity)) == needle {
			name := strings.TrimSpace(pub.Name)
			if name == "" {
				name = pub.Identity
			}
			return name, true
		}
	}
	return "", false
}

func CanonicalDirPayload(dir string) ([]byte, error) {
	return canonicalDirPayload(dir)
}

func canonicalDirPayload(dir string) ([]byte, error) {
	var names []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if rel == SignatureFileName || strings.HasSuffix(rel, "/"+SignatureFileName) {
			return nil
		}
		names = append(names, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, name := range names {
		data, readErr := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name))) // #nosec G304 -- under validated plugin dir
		if readErr != nil {
			_ = w.Close() // #nosec G104 -- cleanup after read failure
			return nil, readErr
		}
		header := &zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		}
		header.SetModTime(fixedZipModTime)
		writer, createErr := w.CreateHeader(header)
		if createErr != nil {
			_ = w.Close() // #nosec G104 -- cleanup after zip header failure
			return nil, createErr
		}
		if _, writeErr := writer.Write(data); writeErr != nil {
			_ = w.Close() // #nosec G104 -- cleanup after zip write failure
			return nil, writeErr
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func CanonicalZipPayload(files []*zip.File) ([]byte, error) {
	return canonicalZipPayload(files)
}

func canonicalZipPayload(files []*zip.File) ([]byte, error) {
	type entry struct {
		name string
		file *zip.File
	}
	var entries []entry
	for _, f := range files {
		clean := filepath.Clean(strings.ReplaceAll(f.Name, `\`, "/"))
		if clean == "" || clean == "." {
			continue
		}
		if strings.HasPrefix(clean, "../") {
			return nil, fmt.Errorf("zip path traversal: %s", f.Name)
		}
		if filepath.Base(clean) == SignatureFileName {
			continue
		}
		if f.FileInfo().IsDir() {
			continue
		}
		entries = append(entries, entry{name: clean, file: f})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, item := range entries {
		rc, err := item.file.Open()
		if err != nil {
			_ = w.Close() // #nosec G104 -- cleanup after zip open failure
			return nil, err
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close() // #nosec G104 -- read complete or error handled below
		if err != nil {
			_ = w.Close() // #nosec G104 -- cleanup after zip read failure
			return nil, err
		}
		header := &zip.FileHeader{
			Name:   item.name,
			Method: zip.Deflate,
		}
		header.SetModTime(fixedZipModTime)
		writer, err := w.CreateHeader(header)
		if err != nil {
			_ = w.Close() // #nosec G104 -- cleanup after zip header failure
			return nil, err
		}
		if _, err := writer.Write(data); err != nil {
			_ = w.Close() // #nosec G104 -- cleanup after zip write failure
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func RequireValidSignature(info SignatureInfo) error {
	if !info.Present {
		return nil
	}
	if info.Valid {
		return nil
	}
	if info.Error != "" {
		return fmt.Errorf("invalid extension signature: %s", info.Error)
	}
	return fmt.Errorf("invalid extension signature")
}

func WriteDirSignature(dir string, rsgData []byte) error {
	if len(rsgData) == 0 {
		return fmt.Errorf("signature is required")
	}
	path := filepath.Join(dir, SignatureFileName)
	return os.WriteFile(path, rsgData, 0o600)
}

func EmbedSignatureInZip(zipPath string, rsgData []byte) error {
	if len(rsgData) == 0 {
		return fmt.Errorf("signature is required")
	}
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	tmpPath := zipPath + ".signed"
	out, err := os.Create(tmpPath) // #nosec G304 -- temp sibling of validated zip path
	if err != nil {
		return err
	}
	writer := zip.NewWriter(out)

	wrote := map[string]struct{}{}
	for _, f := range reader.File {
		clean := filepath.Clean(strings.ReplaceAll(f.Name, `\`, "/"))
		if clean == "" || clean == "." {
			continue
		}
		if filepath.Base(clean) == SignatureFileName {
			continue
		}
		if err := copyZipEntry(writer, f, clean); err != nil {
			_ = writer.Close()     // #nosec G104 -- cleanup after zip copy failure
			_ = out.Close()        // #nosec G104 -- cleanup after zip copy failure
			_ = os.Remove(tmpPath) // #nosec G104 -- cleanup after zip copy failure
			return err
		}
		wrote[clean] = struct{}{}
	}
	header := &zip.FileHeader{
		Name:   SignatureFileName,
		Method: zip.Deflate,
	}
	header.SetModTime(fixedZipModTime)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		_ = writer.Close()     // #nosec G104 -- cleanup after zip header failure
		_ = out.Close()        // #nosec G104 -- cleanup after zip header failure
		_ = os.Remove(tmpPath) // #nosec G104 -- cleanup after zip header failure
		return err
	}
	if _, err := entry.Write(rsgData); err != nil {
		_ = writer.Close()     // #nosec G104 -- cleanup after signature write failure
		_ = out.Close()        // #nosec G104 -- cleanup after signature write failure
		_ = os.Remove(tmpPath) // #nosec G104 -- cleanup after signature write failure
		return err
	}
	if err := writer.Close(); err != nil {
		_ = out.Close()        // #nosec G104 -- cleanup after zip close failure
		_ = os.Remove(tmpPath) // #nosec G104 -- cleanup after zip close failure
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmpPath) // #nosec G104 -- cleanup after file close failure
		return err
	}
	return os.Rename(tmpPath, zipPath)
}

func copyZipEntry(writer *zip.Writer, f *zip.File, name string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	}
	header.SetModTime(fixedZipModTime)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = entry.Write(data)
	return err
}

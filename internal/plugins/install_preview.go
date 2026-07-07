// SPDX-License-Identifier: MIT
package plugins

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type InstallPreview struct {
	Manifest             Manifest           `json:"manifest"`
	NetworkEndpoints     []string           `json:"networkEndpoints"`
	RequiresNetworkFetch bool               `json:"requiresNetworkFetch"`
	Signature            SignatureInfo      `json:"signature"`
	Security             SecurityAssessment `json:"security"`
	I18nLocales          []string           `json:"i18nLocales"`
}

func BuildInstallPreview(manifest Manifest, dir string, embedded map[string][]byte, signature SignatureInfo) InstallPreview {
	return InstallPreview{
		Manifest:             manifest,
		NetworkEndpoints:     CollectNetworkEndpoints(manifest, dir, embedded),
		RequiresNetworkFetch: HasPermission(manifest, PermNetworkFetch),
		Signature:            signature,
		Security:             AssessExtension(manifest, dir, embedded, signature),
		I18nLocales:          CollectPluginI18nLocales(dir, embedded),
	}
}

func PreviewInstallFromDir(dir string) (InstallPreview, error) {
	manifest, err := LoadManifest(dir)
	if err != nil {
		return InstallPreview{}, err
	}
	if err := ValidatePermissions(manifest.Permissions); err != nil {
		return InstallPreview{}, err
	}
	return BuildInstallPreview(manifest, dir, nil, VerifyDirSignature(dir)), nil
}

func PreviewInstallFromZip(zipPath string) (InstallPreview, error) {
	manifest, tmpDir, err := extractZipManifest(zipPath)
	if err != nil {
		return InstallPreview{}, err
	}
	defer os.RemoveAll(tmpDir)
	if err := ValidatePermissions(manifest.Permissions); err != nil {
		return InstallPreview{}, err
	}
	return BuildInstallPreview(manifest, tmpDir, nil, VerifyZipSignature(zipPath)), nil
}

func PreviewInstallFromWasm(wasmPath string) (InstallPreview, error) {
	info, err := os.Stat(wasmPath)
	if err != nil {
		return InstallPreview{}, err
	}
	if info.Size() > maxWasmBundleBytes {
		return InstallPreview{}, fmt.Errorf("wasm module too large")
	}
	data, err := os.ReadFile(wasmPath) // #nosec G304 -- path from desktop file picker
	if err != nil {
		return InstallPreview{}, err
	}
	bundle, err := ParseWasmBundle(data)
	if err != nil {
		return InstallPreview{}, err
	}
	if err := bundle.ValidateEmbedded(); err != nil {
		return InstallPreview{}, err
	}
	if err := ValidatePermissions(bundle.Manifest.Permissions); err != nil {
		return InstallPreview{}, err
	}
	embedded := make(map[string][]byte, len(bundle.Files))
	for path, content := range bundle.Files {
		embedded[path] = []byte(content)
	}
	return BuildInstallPreview(bundle.Manifest, "", embedded, VerifyWasmSignature(data)), nil
}

func extractZipManifest(zipPath string) (Manifest, string, error) {
	info, err := os.Stat(zipPath)
	if err != nil {
		return Manifest{}, "", err
	}
	if info.Size() > maxZipBytes {
		return Manifest{}, "", fmt.Errorf("zip file too large")
	}
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return Manifest{}, "", err
	}
	defer reader.Close()
	if len(reader.File) > maxZipFiles {
		return Manifest{}, "", fmt.Errorf("zip has too many files")
	}
	var total int64
	var manifestData []byte
	tmpDir, err := os.MkdirTemp("", "renplugin-preview-*")
	if err != nil {
		return Manifest{}, "", err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }
	for _, f := range reader.File {
		if f.UncompressedSize64 > uint64(maxZipUncompressed) {
			cleanup()
			return Manifest{}, "", fmt.Errorf("zip uncompressed size too large")
		}
		total += int64(f.UncompressedSize64) // #nosec G115 -- bounded by check above
		if total > maxZipUncompressed {
			cleanup()
			return Manifest{}, "", fmt.Errorf("zip uncompressed size too large")
		}
		clean := filepath.Clean(strings.ReplaceAll(f.Name, `\`, "/"))
		if clean == "" || clean == "." {
			continue
		}
		dest, err := safeZipJoin(tmpDir, clean)
		if err != nil {
			cleanup()
			return Manifest{}, "", err
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, 0o750); err != nil {
				cleanup()
				return Manifest{}, "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
			cleanup()
			return Manifest{}, "", err
		}
		if err := extractZipFile(f, dest); err != nil {
			cleanup()
			return Manifest{}, "", err
		}
		if filepath.Base(clean) == ManifestFileName && manifestData == nil {
			manifestData, _ = os.ReadFile(dest) // #nosec G304 -- path cleaned and under temp extract dir
		}
	}
	if len(manifestData) == 0 {
		cleanup()
		return Manifest{}, "", fmt.Errorf("manifest not found in zip")
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		cleanup()
		return Manifest{}, "", err
	}
	if err := ValidateManifest(manifest); err != nil {
		cleanup()
		return Manifest{}, "", err
	}
	return manifest, tmpDir, nil
}

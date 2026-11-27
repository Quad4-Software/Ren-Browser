// SPDX-License-Identifier: MIT
package plugins

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxZipBytes        = 32 * 1024 * 1024
	maxZipFiles        = 256
	maxZipUncompressed = 64 * 1024 * 1024
)

func safeZipJoin(root, name string) (string, error) {
	rootClean := filepath.Clean(root)
	clean := filepath.Clean(strings.ReplaceAll(name, `\`, "/"))
	if clean == "" || clean == "." {
		return rootClean, nil
	}
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("zip path traversal: %s", name)
	}
	target := filepath.Join(rootClean, clean)
	rel, err := filepath.Rel(rootClean, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("zip path traversal: %s", name)
	}
	return target, nil
}

func extractZipFile(f *zip.File, dest string) error {
	if !f.FileInfo().Mode().IsRegular() {
		return fmt.Errorf("zip entry is not a regular file: %s", f.Name)
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	limit := int64(f.UncompressedSize64) + 1 // #nosec G115 -- caller bounds this at maxZipUncompressed
	written, copyErr := io.Copy(out, io.LimitReader(rc, limit))
	if copyErr != nil {
		return copyErr
	}
	if written > int64(f.UncompressedSize64) { // #nosec G115 -- caller bounds this at maxZipUncompressed
		return fmt.Errorf("zip entry exceeds declared size: %s", f.Name)
	}
	return nil
}

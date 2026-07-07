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
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()) // #nosec G304
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc) // #nosec G110 -- zip size limits enforced before extract
	return err
}

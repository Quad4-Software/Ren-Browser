// SPDX-License-Identifier: MIT
package rns

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"quad4/reticulum-go/pkg/identity"
)

func openStorageRoot(storageDir string) (*os.Root, error) {
	return os.OpenRoot(storageDir)
}

func openIdentitiesRoot(storageDir string) (*os.Root, error) {
	if err := os.MkdirAll(identitiesDir(storageDir), 0o700); err != nil {
		return nil, fmt.Errorf("identities dir: %w", err)
	}
	return os.OpenRoot(identitiesDir(storageDir))
}

func readStorageFile(storageDir, name string) ([]byte, error) {
	root, err := openStorageRoot(storageDir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return root.ReadFile(name)
}

func atomicWriteRootFile(root *os.Root, name string, data []byte, perm os.FileMode) error {
	tmp := name + ".tmp"
	if err := root.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	if err := root.Rename(tmp, name); err != nil {
		_ = root.Remove(tmp)
		return err
	}
	return nil
}

func atomicWriteStorageFile(storageDir, name string, data []byte, perm os.FileMode) error {
	root, err := openStorageRoot(storageDir)
	if err != nil {
		return err
	}
	defer root.Close()
	return atomicWriteRootFile(root, name, data, perm)
}

func readIdentityKeyBytes(storageDir, id string) ([]byte, error) {
	if err := ensureIdentityKeyUnderDir(storageDir, id); err != nil {
		return nil, err
	}
	root, err := openIdentitiesRoot(storageDir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	data, err := root.ReadFile(id)
	if err != nil {
		return nil, err
	}
	if err := validateIdentityKeyData(data); err != nil {
		return nil, err
	}
	return data, nil
}

func loadIdentityFromStorage(storageDir, id string) (*identity.Identity, error) {
	data, err := readIdentityKeyBytes(storageDir, id)
	if err != nil {
		return nil, err
	}
	return identity.FromBytes(data)
}

func writeIdentityToStorage(storageDir, id string, ident *identity.Identity) error {
	if ident == nil {
		return fmt.Errorf("nil identity")
	}
	if err := ensureIdentityKeyUnderDir(storageDir, id); err != nil {
		return err
	}
	privateKey, err := ident.GetPrivateKey()
	if err != nil {
		return err
	}
	root, err := openIdentitiesRoot(storageDir)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.WriteFile(id, privateKey, 0o600)
}

func removeIdentityKeyStorage(storageDir, id string) error {
	if err := ensureIdentityKeyUnderDir(storageDir, id); err != nil {
		return err
	}
	root, err := openIdentitiesRoot(storageDir)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Remove(id); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func atomicWriteExportFile(destPath string, data []byte) error {
	destPath = filepath.Clean(destPath)
	if destPath == "" || destPath == "." {
		return fmt.Errorf("export path is required")
	}
	dir := filepath.Dir(destPath)
	name := filepath.Base(destPath)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return fmt.Errorf("export path must include a file name")
	}
	if strings.Contains(name, string(filepath.Separator)) {
		return fmt.Errorf("export path must include a file name")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("prepare export directory: %w", err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return fmt.Errorf("open export directory: %w", err)
	}
	defer root.Close()
	return atomicWriteRootFile(root, name, data, 0o600)
}

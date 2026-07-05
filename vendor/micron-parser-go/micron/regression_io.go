// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"io/fs"
	"os"
	"strings"
)

const regressionDataDir = "testdata/regressions"

func withRegressionRoot(fn func(root *os.Root) error) error {
	root, err := os.OpenRoot(regressionDataDir)
	if err != nil {
		return err
	}
	defer root.Close()
	return fn(root)
}

func readRegressionFile(root *os.Root, name string) ([]byte, error) {
	return root.ReadFile(name)
}

func listRegressionMuNames(root *os.Root) ([]string, error) {
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".mu") {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}

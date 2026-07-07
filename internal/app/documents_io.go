// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"os"
)

func readFileLimited(path string, max int) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory")
	}
	if max > 0 && info.Size() > int64(max) {
		return nil, fmt.Errorf("document too large to view in browser (max %d bytes)", max)
	}
	data, err := os.ReadFile(path) // #nosec G304 -- path validated under download dir
	if err != nil {
		return nil, err
	}
	if max > 0 && len(data) > max {
		return nil, fmt.Errorf("document too large to view in browser (max %d bytes)", max)
	}
	return data, nil
}

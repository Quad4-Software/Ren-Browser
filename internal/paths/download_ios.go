//go:build ios

// SPDX-License-Identifier: MIT

package paths

import "path/filepath"

// UserDownloadDir returns a writable downloads path under the app Documents
// directory. The container-root Downloads folder is not creatable on iOS.
func UserDownloadDir() string {
	return filepath.Join(DataRoot(), "Downloads")
}

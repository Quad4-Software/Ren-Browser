//go:build !android && !ios

// SPDX-License-Identifier: MIT

package paths

func UserDownloadDir() string {
	return ""
}

// SPDX-License-Identifier: MIT
package limits

import (
	"os"
	"strconv"
	"strings"
)

const (
	DefaultMaxPageBytes     = 8 * 1024 * 1024
	DefaultMaxFileBytes     = 128 * 1024 * 1024
	DefaultMaxAssetBytes    = 32 * 1024 * 1024
	DefaultMaxTabFieldBytes = 256 * 1024
)

func MaxPageBytes() int {
	return envBytes("REN_BROWSER_MAX_PAGE_BYTES", DefaultMaxPageBytes)
}

func MaxFileBytes() int {
	return envBytes("REN_BROWSER_MAX_FILE_BYTES", DefaultMaxFileBytes)
}

func MaxAssetBytes() int {
	return envBytes("REN_BROWSER_MAX_ASSET_BYTES", DefaultMaxAssetBytes)
}

func MaxTabFieldBytes() int {
	return envBytes("REN_BROWSER_MAX_TAB_FIELD_BYTES", DefaultMaxTabFieldBytes)
}

func MaxFetchBytes(path string) int {
	if strings.HasPrefix(path, "/file/") {
		return MaxFileBytes()
	}
	return MaxPageBytes()
}

func envBytes(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func TruncateString(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}

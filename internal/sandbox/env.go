// SPDX-License-Identifier: MIT
package sandbox

import (
	"os"
	"strings"

	"renbrowser/internal/brand"
)

func envLandlockOverride() *bool {
	return envBoolOverride(brand.EnvPrefix + "_LANDLOCK")
}

func envSeccompOverride() *bool {
	return envBoolOverride(brand.EnvPrefix + "_SECCOMP")
}

func envBoolOverride(key string) *bool {
	raw := os.Getenv(key)
	if raw == "" {
		return nil
	}
	val := strings.ToLower(strings.TrimSpace(raw))
	switch val {
	case "0", "false", "no", "off":
		v := false
		return &v
	case "1", "true", "yes", "on":
		v := true
		return &v
	default:
		return nil
	}
}

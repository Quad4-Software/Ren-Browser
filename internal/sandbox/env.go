// SPDX-License-Identifier: MIT
package sandbox

import (
	"os"
	"strings"

	"renbrowser/internal/brand"
)

func envLandlockOverride() *bool {
	raw := os.Getenv(brand.EnvPrefix + "_LANDLOCK")
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

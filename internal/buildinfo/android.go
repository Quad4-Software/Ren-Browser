// SPDX-License-Identifier: MIT
package buildinfo

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	androidVersionCodeMajor   = 10000
	androidVersionCodeMinor   = 100
	androidVersionCodeMaxPart = 99
)

// AndroidVersionCode maps a semver string to the integer Android versionCode.
// Package Manager uses this integer to decide whether an APK is an update.
// Formula is major*10000 + minor*100 + patch. Prerelease and build suffixes
// are ignored. v0.2.1 and v0.2.2 were both shipped with versionCode 2, so
// 0.2.x values from this function stay above 2 (0.2.1 yields 201).
func AndroidVersionCode(version string) (int, error) {
	core := strings.TrimSpace(version)
	core = strings.TrimPrefix(core, "v")
	core = strings.TrimPrefix(core, "V")
	if i := strings.IndexByte(core, '+'); i >= 0 {
		core = core[:i]
	}
	if i := strings.IndexByte(core, '-'); i >= 0 {
		core = core[:i]
	}
	parts := strings.Split(core, ".")
	if len(parts) < 3 {
		return 0, fmt.Errorf("version %q is not major.minor.patch", version)
	}
	major, err := parseVersionPart(parts[0], "major", version)
	if err != nil {
		return 0, err
	}
	minor, err := parseVersionPart(parts[1], "minor", version)
	if err != nil {
		return 0, err
	}
	patch, err := parseVersionPart(parts[2], "patch", version)
	if err != nil {
		return 0, err
	}
	if minor > androidVersionCodeMaxPart || patch > androidVersionCodeMaxPart {
		return 0, fmt.Errorf("version %q overflows android versionCode layout", version)
	}
	return major*androidVersionCodeMajor + minor*androidVersionCodeMinor + patch, nil
}

func parseVersionPart(raw, name, version string) (int, error) {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("version %q has invalid %s", version, name)
	}
	return n, nil
}

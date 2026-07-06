// SPDX-License-Identifier: MIT
package buildinfo

import (
	_ "embed"
	"runtime/debug"
	"strings"
)

//go:embed reticulum_version.txt
var reticulumVersionPinned string

func ReticulumGoVersion() string {
	if v := moduleVersion("quad4/reticulum-go"); v != "" && !isPseudoModuleVersion(v) {
		return normalizeVersionTag(v)
	}
	if v := strings.TrimSpace(reticulumVersionPinned); v != "" {
		return normalizeVersionTag(v)
	}
	return "unknown"
}

func moduleVersion(path string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, dep := range info.Deps {
		if dep == nil || dep.Path != path {
			continue
		}
		if dep.Replace != nil && strings.TrimSpace(dep.Replace.Version) != "" {
			return dep.Replace.Version
		}
		return dep.Version
	}
	return ""
}

func isPseudoModuleVersion(version string) bool {
	version = strings.TrimSpace(version)
	return version == "" || version == "v0.0.0" || strings.HasPrefix(version, "v0.0.0-")
}

func normalizeVersionTag(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}

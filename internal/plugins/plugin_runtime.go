// SPDX-License-Identifier: MIT
package plugins

import (
	"encoding/json"
	"fmt"
	"slices"
)

type RuntimeSettings struct {
	GrantedPermissions []string `json:"grantedPermissions,omitempty"`
	IntegrityHash      string   `json:"integrityHash,omitempty"`
	Tampered           bool     `json:"tampered,omitempty"`
}

func ParseRuntimeSettings(raw string) RuntimeSettings {
	if raw == "" {
		return RuntimeSettings{}
	}
	var settings RuntimeSettings
	_ = json.Unmarshal([]byte(raw), &settings)
	return settings
}

func (s RuntimeSettings) JSON() string {
	raw, err := json.Marshal(s)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func DefaultGrantedPermissions(manifest Manifest) []string {
	if len(manifest.Permissions) == 0 {
		return nil
	}
	out := make([]string, len(manifest.Permissions))
	copy(out, manifest.Permissions)
	return out
}

func NormalizeGrantedPermissions(manifest Manifest, granted []string) []string {
	if len(granted) == 0 {
		return DefaultGrantedPermissions(manifest)
	}
	seen := make(map[string]struct{}, len(granted))
	var out []string
	for _, perm := range granted {
		if !HasPermission(manifest, perm) {
			continue
		}
		if _, ok := seen[perm]; ok {
			continue
		}
		seen[perm] = struct{}{}
		out = append(out, perm)
	}
	slices.Sort(out)
	return out
}

func HasGrantedPermission(granted []string, perm string) bool {
	return slices.Contains(granted, perm)
}

func RequireGrantedPermission(granted []string, manifest Manifest, perm string) error {
	if !HasPermission(manifest, perm) {
		return fmt.Errorf("plugin %s lacks permission %s", manifest.ID, perm)
	}
	if !HasGrantedPermission(granted, perm) {
		return fmt.Errorf("plugin %s permission %s not granted", manifest.ID, perm)
	}
	return nil
}

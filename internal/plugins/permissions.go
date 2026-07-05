// SPDX-License-Identifier: MIT
package plugins

import "fmt"

const (
	PermStoragePlugin     = "storage.plugin"
	PermNavigationRead    = "navigation.read"
	PermNavigationWrite   = "navigation.write"
	PermNetworkFetch      = "network.fetch"
	PermEventsEmit        = "events.emit"
	PermEventsSubscribe   = "events.subscribe"
	PermDevtoolsNetwork   = "devtools.network"
	PermRenderUnsanitized = "render.unsanitized"
)

var knownPermissions = map[string]struct{}{
	PermStoragePlugin:     {},
	PermNavigationRead:    {},
	PermNavigationWrite:   {},
	PermNetworkFetch:      {},
	PermEventsEmit:        {},
	PermEventsSubscribe:   {},
	PermDevtoolsNetwork:   {},
	PermRenderUnsanitized: {},
}

func ValidatePermissions(perms []string) error {
	for _, p := range perms {
		if _, ok := knownPermissions[p]; !ok {
			return fmt.Errorf("unknown permission %q", p)
		}
	}
	return nil
}

func HasPermission(m Manifest, perm string) bool {
	for _, p := range m.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

func RequirePermission(m Manifest, perm string) error {
	if HasPermission(m, perm) {
		return nil
	}
	return fmt.Errorf("plugin %s lacks permission %s", m.ID, perm)
}

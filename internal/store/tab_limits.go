// SPDX-License-Identifier: MIT
package store

import "renbrowser/internal/limits"

func clampTabSnapshots(tabs []TabSnapshot) []TabSnapshot {
	max := limits.MaxTabFieldBytes()
	if max <= 0 {
		return tabs
	}
	out := make([]TabSnapshot, len(tabs))
	for i, tab := range tabs {
		out[i] = tab
		out[i].HTML = limits.TruncateString(tab.HTML, max)
		out[i].LastRaw = limits.TruncateString(tab.LastRaw, max)
	}
	return out
}

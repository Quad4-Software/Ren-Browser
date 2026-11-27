// SPDX-License-Identifier: MIT
package store

import "renbrowser/internal/limits"

func clampTabSnapshots(tabs []TabSnapshot) []TabSnapshot {
	bodyMax := limits.MaxPageBytes()
	fieldMax := limits.MaxTabFieldBytes()
	out := make([]TabSnapshot, len(tabs))
	for i, tab := range tabs {
		out[i] = tab
		if bodyMax > 0 {
			out[i].HTML = limits.TruncateString(tab.HTML, bodyMax)
			out[i].LastRaw = limits.TruncateString(tab.LastRaw, bodyMax)
		}
		if fieldMax > 0 {
			out[i].ID = limits.TruncateString(tab.ID, fieldMax)
			out[i].Title = limits.TruncateString(tab.Title, fieldMax)
			out[i].URL = limits.TruncateString(tab.URL, fieldMax)
			out[i].GroupID = limits.TruncateString(tab.GroupID, fieldMax)
			out[i].ContentType = limits.TruncateString(tab.ContentType, fieldMax)
			out[i].Error = limits.TruncateString(tab.Error, fieldMax)
			out[i].PageFG = limits.TruncateString(tab.PageFG, fieldMax)
			out[i].PageBG = limits.TruncateString(tab.PageBG, fieldMax)
		}
	}
	return out
}

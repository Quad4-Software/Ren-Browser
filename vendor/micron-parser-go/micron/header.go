// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

// PageColors holds optional page-level colors from leading #!fg= / #!bg=
// directives (three- or six-digit hex when valid).
type PageColors struct {
	FG         string `json:"fg"`
	BG         string `json:"bg"`
	FoldOpen   string `json:"fold_open"`
	FoldClosed string `json:"fold_closed"`
}

// ParseHeaderTags reads leading #!fg= and #!bg= lines at the start of markup,
// stopping at the first non-directive line. Those lines stay in the markup string.
// ConvertMicronToHTML applies the same rules when rendering.
func ParseHeaderTags(markup string) PageColors {
	var out PageColors
	lines := strings.SplitSeq(markup, "\n")
	for line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if !strings.HasPrefix(t, "#!") {
			break
		}
		if strings.HasPrefix(t, "#!fg=") {
			c := strings.TrimSpace(t[5:])
			if len(c) == 3 || len(c) == 6 {
				out.FG = c
			}
			continue
		}
		if strings.HasPrefix(t, "#!bg=") {
			c := strings.TrimSpace(t[5:])
			if len(c) == 3 || len(c) == 6 {
				out.BG = c
			}
		}
		if strings.HasPrefix(t, "#!fold") {
			fields := strings.Fields(t[6:])
			if len(fields) >= 2 {
				out.FoldOpen = fields[0]
				out.FoldClosed = fields[1]
			} else if len(fields) == 1 {
				out.FoldOpen = fields[0]
				out.FoldClosed = fields[0]
			}
		}
	}
	return out
}

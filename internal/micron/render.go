// SPDX-License-Identifier: MIT
package micron

import (
	mp "micron-parser-go/micron"
)

func ToHTMLDark(source string) string {
	html, _, _ := RenderDark(source)
	return html
}

func RenderDark(source string) (html, fg, bg string) {
	pc := mp.ParseHeaderTags(source)
	// Always emit Mu-mnt cells so ASCII/box art stays column-aligned.
	// MicronPreserveLayout is CSS-only (horizontal scroll) and must not gate this.
	p := mp.Parser{
		DarkTheme:      true,
		ForceMonospace: true,
	}
	html = p.ConvertMicronToHTML(source)
	return html, pc.FG, pc.BG
}

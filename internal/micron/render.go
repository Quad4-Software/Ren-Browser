// SPDX-License-Identifier: MIT
package micron

import (
	mp "micron-parser-go/micron"
)

type Options struct {
	DarkTheme      bool
	ForceMonospace bool
}

func ToHTML(source string, opts Options) string {
	p := mp.Parser{
		DarkTheme:      opts.DarkTheme,
		ForceMonospace: opts.ForceMonospace,
	}
	return p.ConvertMicronToHTML(source)
}

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

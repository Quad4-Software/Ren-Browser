// SPDX-License-Identifier: MIT
package micron

import (
	mp "micron-parser-go/micron"
)

type Options struct {
	DarkTheme      bool
	ForceMonospace bool
}

// GetForceMonospace is a global hook to dynamically retrieve the ForceMonospace setting.
var GetForceMonospace func() bool

var darkParser = mp.Parser{DarkTheme: true}

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
	force := false
	if GetForceMonospace != nil {
		force = GetForceMonospace()
	}
	p := mp.Parser{
		DarkTheme:      true,
		ForceMonospace: force,
	}
	html = p.ConvertMicronToHTML(source)
	return html, pc.FG, pc.BG
}

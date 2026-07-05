// SPDX-License-Identifier: MIT
package content

import (
	"fmt"
	"html"
)

type AboutInfo struct {
	AppName         string
	Version         string
	Build           string
	License         string
	GoVersion       string
	OS              string
	Arch            string
	WailsVersion    string
	ReticulumConfig string
	DataPath        string
}

func RenderAbout(info AboutInfo) string {
	rows := []struct {
		label string
		value string
	}{
		{"Application", info.AppName},
		{"Version", info.Version},
		{"Build", info.Build},
		{"License", info.License},
		{"Go", info.GoVersion},
		{"Platform", fmt.Sprintf("%s/%s", info.OS, info.Arch)},
		{"Wails", info.WailsVersion},
		{"Reticulum config", info.ReticulumConfig},
		{"Data", info.DataPath},
	}

	var body string
	for _, row := range rows {
		body += fmt.Sprintf(
			`<tr><th>%s</th><td>%s</td></tr>`,
			html.EscapeString(row.label),
			html.EscapeString(row.value),
		)
	}

	return `<article class="about-page">` +
		`<h1>` + html.EscapeString(info.AppName) + `</h1>` +
		`<p class="about-tagline">Reticulum browser for NomadNet pages.</p>` +
		`<table class="about-table"><tbody>` + body + `</tbody></table>` +
		`<p class="about-hint">Type a NomadNet URL in the address bar or open Discovery to browse the mesh. ` +
		`<a href="license:">View license</a>.</p>` +
		`</article>`
}

package content

import (
	_ "embed"
	"html"
	"strings"
)

//go:embed LICENSE
var licenseText string

const ProjectLicense = "MIT"

func RenderLicense() string {
	body := html.EscapeString(strings.TrimSpace(licenseText))
	return `<article class="license-page">` +
		`<h1>License</h1>` +
		`<p class="license-spdx"><code>SPDX-License-Identifier: MIT</code></p>` +
		`<pre class="license-text">` + body + `</pre>` +
		`<p class="license-hint"><a href="about:">About</a> · Type <code>license</code> in the address bar to open this page.</p>` +
		`</article>`
}

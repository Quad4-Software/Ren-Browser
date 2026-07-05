// SPDX-License-Identifier: MIT
package plugins

var SanitizeHTML func(string) string

func SanitizePluginHTML(html string, manifest Manifest) string {
	if HasPermission(manifest, PermRenderUnsanitized) {
		return html
	}
	if SanitizeHTML != nil {
		return SanitizeHTML(html)
	}
	return html
}

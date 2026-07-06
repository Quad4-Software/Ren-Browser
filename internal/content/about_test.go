// SPDX-License-Identifier: MIT
package content

import "strings"
import "testing"

func TestRenderAboutIncludesRuntime(t *testing.T) {
	html := RenderAbout(AboutInfo{
		AppName: "RenBrowser",
		Version: "1.0.0",
		Runtime: []AboutRow{
			{Label: "GTK", Value: "4.14.0"},
			{Label: "WebKitGTK", Value: "2.44.0"},
		},
	})
	if !strings.Contains(html, "GTK") || !strings.Contains(html, "4.14.0") {
		t.Fatalf("missing runtime rows: %s", html)
	}
	if !strings.Contains(html, "WebKitGTK") || !strings.Contains(html, "2.44.0") {
		t.Fatalf("missing webkit row: %s", html)
	}
}

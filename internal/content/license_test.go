package content_test

import (
	"strings"
	"testing"

	"renbrowser/internal/content"
)

func TestRenderLicense(t *testing.T) {
	html := content.RenderLicense()
	if !strings.Contains(html, `class="license-page"`) {
		t.Fatal("expected license page markup")
	}
	if !strings.Contains(html, "MIT License") {
		t.Fatal("expected MIT license text")
	}
	if !strings.Contains(html, "SPDX-License-Identifier: MIT") {
		t.Fatal("expected SPDX identifier")
	}
}

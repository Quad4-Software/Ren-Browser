// SPDX-License-Identifier: MIT

package qr_test

import (
	"strings"
	"testing"

	"renbrowser/internal/qr"
)

func TestEncodeSVGSize(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"hello", 21},
		{"http://192.168.1.5:8080/renbrowser-1.0.0.apk", 33},
	}
	for _, tc := range tests {
		size, err := qr.Size(tc.text)
		if err != nil {
			t.Fatalf("Size(%q): %v", tc.text, err)
		}
		if size != tc.want {
			t.Fatalf("Size(%q) = %d, want %d", tc.text, size, tc.want)
		}
		svg, err := qr.EncodeSVG(tc.text, 4)
		if err != nil {
			t.Fatalf("EncodeSVG(%q): %v", tc.text, err)
		}
		if !strings.Contains(svg, "<svg") || !strings.Contains(svg, `fill="#000"`) {
			t.Fatalf("EncodeSVG(%q) missing expected svg content", tc.text)
		}
	}
}

func TestDataURL(t *testing.T) {
	url, err := qr.DataURL("renbrowser", 4)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "data:image/svg+xml;base64,") {
		t.Fatalf("unexpected data url prefix: %s", url[:32])
	}
}

func TestEmptyRejected(t *testing.T) {
	if _, err := qr.EncodeSVG("   ", 4); err == nil {
		t.Fatal("expected error for empty text")
	}
}

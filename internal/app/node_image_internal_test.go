// SPDX-License-Identifier: MIT
package app

import "testing"

func TestValidateNodeImagePath(t *testing.T) {
	valid := []string{
		"/media/harbour.png",
		"/media/x.WEBP",
		"/file/sticker.webp",
		"/media/a/b/c.tiff",
		"media/x.jpeg",
	}
	for _, p := range valid {
		if err := validateNodeImagePath(p); err != nil {
			t.Fatalf("%q should be valid: %v", p, err)
		}
	}

	invalid := []string{
		"",
		"/page/index.mu",
		"/media/x.svg",
		"/media/x.exe",
		"/file/x.png",
		"/file/x.jpg",
		"/media/../secret.png",
		"/media/x.exe`img=1",
		"/media/evil|x.png",
		"/media/con\x01trol.png",
	}
	for _, p := range invalid {
		if err := validateNodeImagePath(p); err == nil {
			t.Fatalf("%q should be rejected", p)
		}
	}
}

func TestSniffNodeImageMime(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0}
	if mime, err := sniffNodeImageMime(png); err != nil || mime != "image/png" {
		t.Fatalf("png sniff: %q %v", mime, err)
	}
	webp := []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")
	if mime, err := sniffNodeImageMime(webp); err != nil || mime != "image/webp" {
		t.Fatalf("webp sniff: %q %v", mime, err)
	}
	if _, err := sniffNodeImageMime([]byte("<svg xmlns=...>")); err == nil {
		t.Fatal("svg must be rejected")
	}
	if _, err := sniffNodeImageMime([]byte("plain text body")); err == nil {
		t.Fatal("text must be rejected")
	}
	if _, err := sniffNodeImageMime(nil); err == nil {
		t.Fatal("empty body must be rejected")
	}
}

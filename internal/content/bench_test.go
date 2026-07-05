package content_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"renbrowser/internal/content"
)

const (
	smallMicron = "`>Title\nplain line\n[`link`:/page/other.mu`label]"
	smallHTML   = "<html><body><p>hello</p></body></html>"
)

var largeHTML = "<html><body>" + strings.Repeat("<p>paragraph</p>", 256) + "</body></html>"

func loadNomadNetGuide(b *testing.B) []byte {
	b.Helper()
	path := filepath.Join("..", "..", "..", "micron-parser-go", "micron", "testdata", "nomadnet_guide.mu")
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skip("nomadnet guide fixture unavailable:", err)
	}
	return data
}

func BenchmarkRenderMicronSmall(b *testing.B) {
	body := []byte(smallMicron)
	b.SetBytes(int64(len(body)))
	b.ReportAllocs()
	for b.Loop() {
		_ = content.Render("/page/index.mu", body, "")
	}
}

func BenchmarkRenderMicronGuide(b *testing.B) {
	body := loadNomadNetGuide(b)
	b.SetBytes(int64(len(body)))
	b.ReportAllocs()
	for b.Loop() {
		_ = content.Render("/page/index.mu", body, "abb3ebcd03cb2388a838e70c001291f9")
	}
}

func BenchmarkRenderHTMLClean(b *testing.B) {
	body := []byte(smallHTML)
	b.SetBytes(int64(len(body)))
	b.ReportAllocs()
	for b.Loop() {
		_ = content.Render("/page/index.html", body, "")
	}
}

func BenchmarkRenderHTMLLargeClean(b *testing.B) {
	body := []byte(largeHTML)
	b.SetBytes(int64(len(body)))
	b.ReportAllocs()
	for b.Loop() {
		_ = content.Render("/page/index.html", body, "")
	}
}

func BenchmarkSanitizeHTMLClean(b *testing.B) {
	b.SetBytes(int64(len(largeHTML)))
	b.ReportAllocs()
	for b.Loop() {
		_ = content.SanitizeHTML(largeHTML)
	}
}

func BenchmarkSanitizeHTMLDirty(b *testing.B) {
	in := `<div onclick="x()"><script>alert(1)</script>` + largeHTML + `</div>`
	b.SetBytes(int64(len(in)))
	b.ReportAllocs()
	for b.Loop() {
		_ = content.SanitizeHTML(in)
	}
}

func BenchmarkRenderMarkdown(b *testing.B) {
	body := []byte("# Title\n\n- one\n- two\n\n```\ncode\n```\n")
	b.SetBytes(int64(len(body)))
	b.ReportAllocs()
	for b.Loop() {
		_ = content.Render("/page/readme.md", body, "")
	}
}

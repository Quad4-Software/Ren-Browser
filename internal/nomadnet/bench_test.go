// SPDX-License-Identifier: MIT
package nomadnet_test

import (
	"testing"

	"renbrowser/internal/nomadnet"
)

const meshURL = "abb3ebcd03cb2388a838e70c001291f9:/page/index.mu"
const meshURLFields = "abb3ebcd03cb2388a838e70c001291f9:/page/form.mu?user=alice&action=go"
const rnsURL = "rns://abb3ebcd03cb2388a838e70c001291f9/page/index.mu"

func BenchmarkParseURLMesh(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = nomadnet.ParseURL(meshURL)
	}
}

func BenchmarkParseURLMeshFields(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = nomadnet.ParseURL(meshURLFields)
	}
}

func BenchmarkParseURLRNS(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = nomadnet.ParseURL(rnsURL)
	}
}

func BenchmarkDetectContentTypeByPath(b *testing.B) {
	body := []byte("ignored")
	b.ReportAllocs()
	for b.Loop() {
		_ = nomadnet.DetectContentType("/page/index.mu", body)
	}
}

func BenchmarkDetectContentTypeProbe(b *testing.B) {
	body := []byte("<!DOCTYPE html><html><body>ok</body></html>")
	b.SetBytes(int64(len(body)))
	b.ReportAllocs()
	for b.Loop() {
		_ = nomadnet.DetectContentType("/page/resource", body)
	}
}

func BenchmarkFormatURLWithFields(b *testing.B) {
	fields := map[string]string{"a": "1", "b": "2", "c": "3"}
	b.ReportAllocs()
	for b.Loop() {
		_ = nomadnet.FormatURLWithFields("abb3ebcd03cb2388a838e70c001291f9", "/page/x.mu", fields)
	}
}

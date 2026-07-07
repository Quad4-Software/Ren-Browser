// SPDX-License-Identifier: MIT
package plugins

import (
	"encoding/json"
	"testing"
)

func FuzzAssessManifestJSON(f *testing.F) {
	f.Add([]byte(`{"manifestVersion":1,"id":"renbrowser.fuzz","name":"Fuzz","version":"1.0.0","permissions":["network.fetch"]}`))
	f.Add([]byte(`{"manifestVersion":1,"id":"","name":"","version":"","permissions":["render.unsanitized","network.fetch"]}`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		var manifest Manifest
		if err := json.Unmarshal(raw, &manifest); err != nil {
			return
		}
		assessment := AssessExtension(manifest, "", nil, SignatureInfo{})
		if assessment.Score < 0 {
			t.Fatalf("negative score")
		}
		switch assessment.RiskLevel {
		case "low", "medium", "high":
		default:
			t.Fatalf("invalid risk level %q", assessment.RiskLevel)
		}
	})
}

func FuzzExtractURLsFromText(f *testing.F) {
	f.Add("https://example.com/path")
	f.Add("prefix https://foo.bar/baz suffix")
	f.Fuzz(func(t *testing.T, text string) {
		urls := extractURLsFromText(text)
		for _, value := range urls {
			if value == "" {
				t.Fatal("empty url")
			}
		}
	})
}

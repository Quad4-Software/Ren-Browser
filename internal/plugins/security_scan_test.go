// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAssessExtensionFlagsUnsignedNetwork(t *testing.T) {
	manifest := Manifest{
		ID:          "renbrowser.test",
		Name:        "Test",
		Version:     "1.0.0",
		Permissions: []string{PermNetworkFetch},
	}
	assessment := AssessExtension(manifest, "", nil, SignatureInfo{})
	found := false
	for _, finding := range assessment.Findings {
		if finding.ID == "unsigned-network" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected unsigned-network finding")
	}
}

func TestAssessExtensionDetectsEval(t *testing.T) {
	dir := t.TempDir()
	manifest := Manifest{
		ID:      "renbrowser.test",
		Name:    "Test",
		Version: "1.0.0",
	}
	if err := os.WriteFile(filepath.Join(dir, "main.js"), []byte(`eval("alert(1)")`), 0o644); err != nil {
		t.Fatal(err)
	}
	assessment := AssessExtension(manifest, dir, nil, SignatureInfo{})
	found := false
	for _, finding := range assessment.Findings {
		if finding.ID == "js-eval" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected js-eval finding")
	}
	if assessment.RiskLevel == "low" {
		t.Fatalf("expected elevated risk, got %q", assessment.RiskLevel)
	}
}

func TestAssessExtensionInvalidSignatureHighRisk(t *testing.T) {
	manifest := Manifest{ID: "renbrowser.test", Name: "Test", Version: "1.0.0"}
	assessment := AssessExtension(manifest, "", nil, SignatureInfo{Present: true, Valid: false})
	if assessment.RiskLevel != "high" {
		t.Fatalf("risk = %q", assessment.RiskLevel)
	}
}

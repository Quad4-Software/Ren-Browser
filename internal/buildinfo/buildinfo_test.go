package buildinfo_test

import (
	"testing"

	"renbrowser/internal/buildinfo"
)

func TestBuildLabelDevDefault(t *testing.T) {
	if buildinfo.BuildLabel() != "dev" {
		t.Fatalf("BuildLabel() = %q, want dev", buildinfo.BuildLabel())
	}
}

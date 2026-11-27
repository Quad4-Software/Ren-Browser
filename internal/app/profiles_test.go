// SPDX-License-Identifier: MIT
package app

import (
	"strings"
	"testing"
)

func TestImportProfileRejectsUnsupportedVersion(t *testing.T) {
	svc := &BrowserService{}
	err := svc.ImportProfile(`{"version":999}`)
	if err == nil {
		t.Fatal("expected unsupported profile version error")
	}
	if !strings.Contains(err.Error(), "unsupported profile version") {
		t.Fatalf("unexpected error: %v", err)
	}
}

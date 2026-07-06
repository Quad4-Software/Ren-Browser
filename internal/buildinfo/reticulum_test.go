// SPDX-License-Identifier: MIT
package buildinfo

import "testing"

func TestReticulumGoVersion(t *testing.T) {
	if got := ReticulumGoVersion(); got == "" || got == "unknown" {
		t.Fatalf("ReticulumGoVersion() = %q", got)
	}
}

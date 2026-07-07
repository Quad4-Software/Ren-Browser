// SPDX-License-Identifier: MIT
package sandbox_test

import (
	"testing"

	"renbrowser/internal/sandbox"
)

func TestApply_NoLandlockFlag(t *testing.T) {
	sandbox.Apply(sandbox.Options{NoLandlock: true})
	st := sandbox.CurrentStatus()
	if st.Enabled {
		t.Fatal("expected sandbox disabled with --no-landlock")
	}
	if st.Reason == "" {
		t.Fatal("expected disable reason")
	}
}

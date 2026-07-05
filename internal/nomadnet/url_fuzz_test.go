// SPDX-License-Identifier: MIT
package nomadnet_test

import (
	"testing"

	"renbrowser/internal/nomadnet"
)

func FuzzParseURL(f *testing.F) {
	f.Add("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu")
	f.Add("rns://ea6a715f814bdc37e56f80c34da6ad51/page/home.mu")
	f.Add("garbage input")

	f.Fuzz(func(t *testing.T, raw string) {
		parsed, err := nomadnet.ParseURL(raw)
		if err != nil {
			return
		}
		if parsed.NodeHash == "" {
			t.Fatalf("empty node hash for %q", raw)
		}
		if parsed.Path == "" {
			t.Fatalf("empty path for %q", raw)
		}
	})
}

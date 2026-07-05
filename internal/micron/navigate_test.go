// SPDX-License-Identifier: MIT
package micron

import "testing"

func TestResolveNavigationSameNode(t *testing.T) {
	current := "abb3ebcd03cb2388a838e70c001291f9:/page/index.mu"
	next, err := ResolveNavigation(current, ":page/submit.mu`action=run", "user", []FieldInput{
		{Type: "text", Name: "user", Value: "alice"},
	})
	if err != nil {
		t.Fatalf("ResolveNavigation: %v", err)
	}
	if next == "" {
		t.Fatal("expected url")
	}
	if next[:32] != "abb3ebcd03cb2388a838e70c001291f9" {
		t.Fatalf("unexpected node hash in %q", next)
	}
}

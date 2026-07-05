package nomadnet

import (
	"strings"
	"testing"

	"quad4/reticulum-go/pkg/identity"
)

func TestIdentifyErrors(t *testing.T) {
	b := NewBrowser(nil, nil)
	id, err := identity.NewIdentity()
	if err != nil {
		t.Fatal(err)
	}

	if err := b.Identify("abb3ebcd03cb2388a838e70c001291f9", nil); err == nil {
		t.Fatal("expected error for nil identity")
	} else if !strings.Contains(err.Error(), "transport identity unavailable") {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := b.Identify("", id); err == nil {
		t.Fatal("expected error for empty hash")
	} else if !strings.Contains(err.Error(), "invalid node hash") {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := b.Identify("abb3ebcd03cb2388a838e70c001291f9", id); err == nil {
		t.Fatal("expected error without active link")
	} else if !strings.Contains(err.Error(), "no active link") {
		t.Fatalf("unexpected error: %v", err)
	}
}

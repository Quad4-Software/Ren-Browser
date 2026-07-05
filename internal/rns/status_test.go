// SPDX-License-Identifier: MIT
package rns

import (
	"testing"

	"quad4/reticulum-go/pkg/common"
)

func TestListInterfacesSortedByName(t *testing.T) {
	stack := &Stack{
		cfg: &common.ReticulumConfig{
			Interfaces: map[string]*common.InterfaceConfig{
				"Zulu":  {Type: "TCPClientInterface", Enabled: true},
				"Alpha": {Type: "TCPClientInterface", Enabled: false},
				"Mango": {Type: "AutoInterface", Enabled: true},
			},
		},
	}
	got := stack.ListInterfaces()
	if len(got) != 3 {
		t.Fatalf("len = %d", len(got))
	}
	want := []string{"Alpha", "Mango", "Zulu"}
	for i, name := range want {
		if got[i].Name != name {
			t.Fatalf("index %d = %q, want %q", i, got[i].Name, name)
		}
	}
}

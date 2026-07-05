// SPDX-License-Identifier: MIT
package rns

import "testing"

func TestIsTCPClientInterface(t *testing.T) {
	cases := []struct {
		item CommunityInterface
		want bool
	}{
		{CommunityInterface{Type: "TCPClientInterface", Config: "x"}, true},
		{CommunityInterface{TypeName: "TCP Client Interface", Config: "x"}, true},
		{CommunityInterface{Type: "TCPServerInterface", TypeName: "TCP Server", Config: "x"}, false},
		{CommunityInterface{TypeName: "UDP Interface", Config: "x"}, false},
		{CommunityInterface{Type: "TCPClientInterface"}, false},
	}
	for _, tc := range cases {
		if got := IsTCPClientInterface(tc.item); got != tc.want {
			t.Fatalf("IsTCPClientInterface(%+v) = %v, want %v", tc.item, got, tc.want)
		}
	}
}

func TestFilterTCPClientInterfaces(t *testing.T) {
	items := []CommunityInterface{
		{Name: "a", Type: "TCPClientInterface", Config: "[[a]]"},
		{Name: "b", Type: "TCPServerInterface", Config: "[[b]]"},
		{Name: "c", TypeName: "TCP Client", Config: "[[c]]"},
		{Name: "d", Type: "TCPClientInterface", Config: ""},
	}
	out := FilterTCPClientInterfaces(items)
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if out[0].Name != "a" || out[1].Name != "c" {
		t.Fatalf("names = %q, %q", out[0].Name, out[1].Name)
	}
}

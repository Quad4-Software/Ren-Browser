// SPDX-License-Identifier: MIT
package rns

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestFilterSeedableInterfaces(t *testing.T) {
	items := []CommunityInterface{
		{Name: "a", Type: "TCPClientInterface", Config: "[[a]]"},
		{Name: "b", Type: "backbone", TypeName: "BackboneInterface", Config: "[[b]]\n  type = BackboneInterface\n  remote = x.example\n  target_port = 4242"},
		{Name: "c", Type: "TCPServerInterface", Config: "[[c]]"},
	}
	out := FilterSeedableInterfaces(items)
	if usesBackboneTCPFallback() {
		if len(out) != 2 {
			t.Fatalf("len = %d, want 2", len(out))
		}
		if out[1].Name != "b" || out[1].TypeName != "TCPClientInterface" {
			t.Fatalf("backbone item = %+v", out[1])
		}
		return
	}
	if len(out) != 1 || out[0].Name != "a" {
		t.Fatalf("desktop seed = %+v", out)
	}
}

func TestLoadBundledCommunityInterfaces(t *testing.T) {
	items, err := loadBundledCommunityInterfaces(map[string]bool{"Beleth RNS Hub": true})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("expected bundled entries")
	}
	found := false
	for _, item := range items {
		if item.Name == "Beleth RNS Hub" && item.Installed {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected installed bundled item")
	}
}

func TestFetchCommunityInterfacesBundledFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	original := communityDirectoryURL
	communityDirectoryURL = server.URL
	t.Cleanup(func() { communityDirectoryURL = original })

	result, err := FetchCommunityInterfaces(map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.FromBundle {
		t.Fatal("expected bundled fallback")
	}
	if len(result.Items) == 0 {
		t.Fatal("expected bundled items")
	}
}

// SPDX-License-Identifier: MIT
package nomadnet_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"quad4/reticulum-go/pkg/identity"

	"renbrowser/internal/nomadnet"
)

func TestParseMeshURL(t *testing.T) {
	parsed, err := nomadnet.ParseURL("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu")
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if parsed.NodeHash != "abb3ebcd03cb2388a838e70c001291f9" {
		t.Fatalf("hash = %q", parsed.NodeHash)
	}
	if parsed.Path != "/page/index.mu" {
		t.Fatalf("path = %q", parsed.Path)
	}
}

func TestParseRNSURL(t *testing.T) {
	parsed, err := nomadnet.ParseURL("rns://ea6a715f814bdc37e56f80c34da6ad51/page/home.mu")
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if parsed.Path != "/page/home.mu" {
		t.Fatalf("path = %q", parsed.Path)
	}
}

func TestParseURLRejectsGarbage(t *testing.T) {
	_, err := nomadnet.ParseURL("not-a-valid-url")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIsNodeDestination(t *testing.T) {
	id, err := identity.NewIdentity()
	if err != nil {
		t.Fatal(err)
	}

	nameHashFull := sha256.Sum256([]byte(nomadnet.AspectKey))
	nameHash10 := nameHashFull[:10]
	identityHash := identity.TruncatedHash(id.GetPublicKey())
	combined := append(append([]byte(nil), nameHash10...), identityHash...)
	expectedFull := sha256.Sum256(combined)
	destHash := expectedFull[:16]

	if !nomadnet.IsNodeDestination(id, destHash) {
		t.Fatal("expected nomadnet node destination match")
	}

	wrong := make([]byte, 16)
	if nomadnet.IsNodeDestination(id, wrong) {
		t.Fatal("expected mismatch for random hash")
	}
}

func TestDetectContentType(t *testing.T) {
	cases := []struct {
		path string
		body string
		want string
	}{
		{"/page/index.mu", "`>Hello", "micron"},
		{"/page/readme.md", "# Title", "markdown"},
		{"/page/site.html", "<html></html>", "html"},
		{"/page/note.txt", "plain", "plaintext"},
	}
	for _, tc := range cases {
		got := nomadnet.DetectContentType(tc.path, []byte(tc.body))
		if got != tc.want {
			t.Fatalf("%s: got %q want %q", tc.path, got, tc.want)
		}
	}
}

func TestDetectContentTypeUppercaseExtension(t *testing.T) {
	got := nomadnet.DetectContentType("/page/index.MU", []byte("ignored"))
	if got != "micron" {
		t.Fatalf("got %q want micron", got)
	}
}

func TestAnnounceHandlerStoresNode(t *testing.T) {
	id, err := identity.NewIdentity()
	if err != nil {
		t.Fatal(err)
	}

	nameHashFull := sha256.Sum256([]byte(nomadnet.AspectKey))
	nameHash10 := nameHashFull[:10]
	identityHash := identity.TruncatedHash(id.GetPublicKey())
	combined := append(append([]byte(nil), nameHash10...), identityHash...)
	expectedFull := sha256.Sum256(combined)
	destHash := expectedFull[:16]

	h := nomadnet.NewAnnounceHandler()
	if err := h.ReceivedAnnounce(destHash, id, []byte("Test Node"), 2); err != nil {
		t.Fatalf("ReceivedAnnounce: %v", err)
	}

	nodes := h.List()
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d; want 1", len(nodes))
	}
	if nodes[0].Name != "Test Node" {
		t.Fatalf("name = %q", nodes[0].Name)
	}
	if nodes[0].Hash != hex.EncodeToString(destHash) {
		t.Fatalf("hash = %q", nodes[0].Hash)
	}

	gotID, ok := h.Identity(nodes[0].Hash)
	if !ok || gotID == nil {
		t.Fatal("identity not stored")
	}
}

package nomadnet

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/identity"
	"quad4/reticulum-go/pkg/transport"
)

func nodeDestHash(t *testing.T, id *identity.Identity) []byte {
	t.Helper()
	nameHashFull := sha256.Sum256([]byte(AspectKey))
	nameHash10 := nameHashFull[:10]
	identityHash := identity.TruncatedHash(id.GetPublicKey())
	combined := append(append([]byte(nil), nameHash10...), identityHash...)
	expectedFull := sha256.Sum256(combined)
	return expectedFull[:16]
}

func TestTransportHopsPrefersPathTable(t *testing.T) {
	tr := transport.NewTransport(&common.ReticulumConfig{EnableTransport: true})
	t.Cleanup(func() { _ = tr.Close() })
	ident, err := identity.NewIdentity()
	if err != nil {
		t.Fatal(err)
	}
	tr.SetIdentity(ident)

	dest := nodeDestHash(t, ident)
	nodeHash := hex.EncodeToString(dest)
	handler := NewAnnounceHandler()
	if err := handler.ReceivedAnnounce(dest, ident, []byte("Node"), 4); err != nil {
		t.Fatalf("ReceivedAnnounce: %v", err)
	}

	if got := transportHops(tr, dest, handler, nodeHash); got != 4 {
		t.Fatalf("announce fallback hops = %d", got)
	}

	tr.UpdatePath(dest, bytes.Repeat([]byte{0x31}, 16), "out", 2)
	if got := transportHops(tr, dest, handler, nodeHash); got != 4 {
		t.Fatalf("expected announce hops without registered iface, got %d", got)
	}
}

func TestTransportHopsUnknown(t *testing.T) {
	if got := transportHops(nil, bytes.Repeat([]byte{0x01}, 16), nil, "01010101010101010101010101010101"); got != -1 {
		t.Fatalf("got %d", got)
	}
}

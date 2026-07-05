package nomadnet

import (
	"errors"
	"fmt"

	"quad4/reticulum-go/pkg/identity"
	rlink "quad4/reticulum-go/pkg/link"
)

func (b *Browser) Identify(nodeHash string, localID *identity.Identity) error {
	if localID == nil {
		return errors.New("transport identity unavailable")
	}
	key := normalizeHash(nodeHash)
	if key == "" {
		return errors.New("invalid node hash")
	}

	b.mu.Lock()
	lnk := b.links[key]
	b.mu.Unlock()

	if lnk == nil || lnk.GetStatus() != rlink.StatusActive {
		return fmt.Errorf("failed to identify: no active link to destination")
	}
	return lnk.Identify(localID)
}

// SPDX-License-Identifier: MIT
package nomadnet_test

import (
	"testing"

	"renbrowser/internal/nomadnet"
)

func TestBrowserCloseEmpty(t *testing.T) {
	b := nomadnet.NewBrowser(nil, nomadnet.NewAnnounceHandler())
	b.Close()
	b.Close()
}

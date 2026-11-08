// SPDX-License-Identifier: MIT
package app

import (
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/rns"
)

// hermeticTestConfig keeps app tests off host networking: a loopback UDP
// interface starts everywhere, unlike AutoInterface which needs a real NIC
// with IPv6 link-local and multicast support (absent on some CI runners).
const hermeticTestConfig = `[reticulum]
  enable_transport = no
  share_instance = no

[interfaces]

  [[loopback]]
    type = UDPInterface
    enabled = yes
    address = 127.0.0.1:0
`

func newTestBrowserService(t *testing.T) *BrowserService {
	t.Helper()
	svc, _ := newTestBrowserServiceIn(t, t.TempDir())
	return svc
}

func newTestBrowserServiceIn(t *testing.T, dir string) (*BrowserService, func()) {
	t.Helper()

	dbPath := filepath.Join(dir, "profile.db")
	cfgPath := filepath.Join(dir, "config")
	if err := os.WriteFile(cfgPath, []byte(hermeticTestConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	stack, err := rns.NewStack(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewBrowserServiceWithOptions(stack, nil, ServiceOptions{
		ProfilePath: dbPath,
	})
	if err != nil {
		_ = stack.Stop()
		t.Fatal(err)
	}

	release := func() {
		_ = svc.Store().Close()
		_ = stack.Stop()
	}
	t.Cleanup(release)
	return svc, release
}

func reopenTestBrowserService(t *testing.T, dbPath, cfgPath string) *BrowserService {
	t.Helper()

	stack, err := rns.NewStack(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := NewBrowserServiceWithOptions(stack, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		_ = stack.Stop()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = svc.Store().Close()
		_ = stack.Stop()
	})
	return svc
}

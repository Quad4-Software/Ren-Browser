// SPDX-License-Identifier: MIT
package app

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/rns"
)

func newTestBrowserService(t *testing.T) *BrowserService {
	t.Helper()
	svc, _ := newTestBrowserServiceIn(t, t.TempDir())
	return svc
}

func newTestBrowserServiceIn(t *testing.T, dir string) (*BrowserService, func()) {
	t.Helper()

	dbPath := filepath.Join(dir, "profile.db")
	cfgPath := filepath.Join(dir, "config")

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

// SPDX-License-Identifier: MIT
package app

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/rns"
)

func newTestBrowserService(t *testing.T) *BrowserService {
	t.Helper()

	dir := t.TempDir()
	stack, err := rns.NewStack(filepath.Join(dir, "config"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stack.Stop() })

	svc, err := NewBrowserServiceWithOptions(stack, nil, ServiceOptions{
		ProfilePath: filepath.Join(dir, "profile.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = svc.Store().Close() })
	return svc
}

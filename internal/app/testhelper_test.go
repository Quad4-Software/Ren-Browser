// SPDX-License-Identifier: MIT
package app_test

import (
	"path/filepath"
	"testing"

	"renbrowser/internal/app"
	"renbrowser/internal/rns"
)

func newTestService(t *testing.T) *app.BrowserService {
	t.Helper()

	root := t.TempDir()
	stack, err := rns.NewStack(filepath.Join(root, "config"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stack.Stop() })

	svc, err := app.NewBrowserServiceWithOptions(stack, nil, app.ServiceOptions{
		ProfilePath: filepath.Join(root, "profile.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = svc.Store().Close() })
	return svc
}

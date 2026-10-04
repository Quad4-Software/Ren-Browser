//go:build interop

// SPDX-License-Identifier: MIT

package interop_test

import (
	"testing"
	"time"
)

// TestSharedInstanceClientRandomFetch attaches to the operator's ~/.reticulum-go
// config as a shared-instance client (expect rnsd or renbrowser-server as owner)
// and fetches /page/index.mu from a few discovered NomadNet nodes.
func TestSharedInstanceClientRandomFetch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped with -short")
	}
	stack := startSharedInstanceStack(t)

	deadline := time.Now().Add(120 * time.Second)
	for time.Now().Before(deadline) {
		if len(stack.Handler().List()) >= 3 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	nodes := stack.Handler().List()
	t.Logf("discovered %d node(s)", len(nodes))
	if len(nodes) == 0 {
		t.Fatal("no nomadnetwork.node announces on shared instance client")
	}

	got, ok := fetchFirstReachable(t, stack.Browser(), nodes, 8, 2)
	if !ok {
		t.Fatal("could not fetch any discovered node as shared-instance client")
	}
	t.Logf(
		"ok: %s (%s) hops=%d iface=%s %d bytes in %dms",
		got.node.Name,
		got.node.Hash,
		got.result.Hops,
		got.result.Interface,
		len(got.result.Body),
		got.result.DurationMs,
	)
}

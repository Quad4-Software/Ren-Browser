//go:build interop

// SPDX-License-Identifier: MIT

package interop_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"renbrowser/internal/nomadnet"
	"renbrowser/internal/rns"
)

func waitForNodes(handler *nomadnet.AnnounceHandler, timeout time.Duration, min int) []nomadnet.Node {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		nodes := handler.List()
		if len(nodes) >= min {
			return nodes
		}
		time.Sleep(2 * time.Second)
	}
	return handler.List()
}

func TestLiveNomadNetFetch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped with -short")
	}

	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	if err := stack.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stack.Stop() }()

	t.Log("waiting for nomadnetwork.node announces...")
	nodes := waitForNodes(stack.Handler(), 90*time.Second, 1)
	if len(nodes) == 0 {
		t.Fatal("no nodes discovered within timeout")
	}
	t.Logf("discovered %d node(s)", len(nodes))

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	const maxNodes = 5
	const retries = 2
	success := 0

	for i, node := range nodes {
		if i >= maxNodes {
			break
		}
		var lastErr string
		ok := false
		for attempt := 0; attempt <= retries; attempt++ {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			fetch := stack.Browser().Fetch(ctx, node.Hash, "/page/index.mu", nomadnet.RequestData{})
			cancel()
			if fetch.Error == "" && len(fetch.Body) > 0 {
				success++
				ok = true
				t.Logf(
					"ok: %s (%s) %d bytes in %dms",
					node.Name,
					node.Hash,
					len(fetch.Body),
					fetch.DurationMs,
				)
				break
			}
			lastErr = fetch.Error
		}
		if !ok {
			t.Logf("unreachable after retries: %s (%s): %s", node.Name, node.Hash, lastErr)
		}
	}

	if success == 0 {
		t.Fatal("could not fetch any discovered node")
	}
}

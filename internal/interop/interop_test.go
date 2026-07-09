//go:build interop

// SPDX-License-Identifier: MIT

package interop_test

import (
	"context"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"quad4/reticulum-go/pkg/reticulumconfig"

	"renbrowser/internal/nomadnet"
	"renbrowser/internal/paths"
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

func waitForOutboundOnlineInterfaces(stack *rns.Stack, timeout time.Duration, min int) []rns.InterfaceInfo {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		online := outboundOnlineInterfaces(stack)
		if len(online) >= min {
			return online
		}
		time.Sleep(2 * time.Second)
	}
	return outboundOnlineInterfaces(stack)
}

func onlineInterfaces(stack *rns.Stack) []rns.InterfaceInfo {
	out := make([]rns.InterfaceInfo, 0)
	for _, iface := range stack.ListInterfaces() {
		if iface.Enabled && iface.Online {
			out = append(out, iface)
		}
	}
	return out
}

func outboundOnlineInterfaces(stack *rns.Stack) []rns.InterfaceInfo {
	out := make([]rns.InterfaceInfo, 0)
	for _, iface := range onlineInterfaces(stack) {
		if isOutboundInterfaceType(iface.Type) {
			out = append(out, iface)
		}
	}
	return out
}

func isOutboundInterfaceType(t string) bool {
	lower := strings.ToLower(strings.TrimSpace(t))
	switch {
	case strings.Contains(lower, "tcpclient"),
		strings.Contains(lower, "backbone"),
		strings.Contains(lower, "pipe"),
		strings.Contains(lower, "i2p"),
		strings.Contains(lower, "websocket"):
		return true
	default:
		return false
	}
}

func ensureInteropCommunityInterfaces(t *testing.T, stack *rns.Stack) []string {
	t.Helper()
	cfg := stack.Config()
	if cfg == nil {
		t.Fatal("config not loaded")
	}

	result, err := rns.FetchCommunityInterfaces(nil)
	if err != nil {
		t.Fatalf("fetch community interfaces: %v", err)
	}
	added := rns.SeedCommunityInterfaces(cfg, result.Items, rns.DefaultCommunityInterfaceCount)
	if iface := cfg.Interfaces["RNS Testnet TCP"]; iface != nil {
		iface.Enabled = true
		if !containsString(added, iface.Name) && iface.Name != "" {
			added = append(added, iface.Name)
		} else if !containsString(added, "RNS Testnet TCP") {
			added = append(added, "RNS Testnet TCP")
		}
	}
	if !rns.ConfigHasOutboundCommunityInterfaces(cfg) {
		t.Fatal("failed to seed any outbound community interfaces")
	}
	if err := reticulumconfig.SaveConfig(cfg); err != nil {
		t.Fatalf("save seeded config: %v", err)
	}
	if err := stack.ApplyConfig(cfg); err != nil {
		t.Fatalf("apply seeded config: %v", err)
	}
	t.Logf("seeded outbound interfaces (%d, fromBundle=%v): %v", len(added), result.FromBundle, added)
	return added
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func TestLiveNomadNetFetch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped with -short")
	}

	// Isolate from the runner's ~/.reticulum-go so CI always starts from a
	// known default config and can seed community uplinks.
	root := t.TempDir()
	paths.SetDataRoot(root)
	t.Cleanup(func() { paths.SetDataRoot("") })

	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	ensureInteropCommunityInterfaces(t, stack)

	if err := stack.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stack.Stop() }()

	t.Logf("config path: %s", stack.ConfigPath())
	t.Logf("data root: %s", filepath.Join(root, ".reticulum-go"))

	t.Log("waiting for outbound TCP/backbone interfaces to come online...")
	outbound := waitForOutboundOnlineInterfaces(stack, 90*time.Second, 1)
	allOnline := onlineInterfaces(stack)
	t.Logf("online interfaces (%d, outbound=%d):", len(allOnline), len(outbound))
	for _, iface := range stack.ListInterfaces() {
		if !iface.Enabled {
			continue
		}
		state := "offline"
		if iface.Online {
			state = "online"
		}
		t.Logf("  %s type=%s %s tx=%d rx=%d", iface.Name, iface.Type, state, iface.TxBytes, iface.RxBytes)
	}
	if len(outbound) == 0 {
		// AutoInterface alone cannot reach the public mesh from CI.
		t.Fatal("no online outbound TCP/backbone interfaces within timeout")
	}

	status := stack.Status()
	t.Logf(
		"status: online=%v transport=%v share=%v mode=%s connectedShared=%v",
		status.Online,
		status.EnableTransport,
		status.ShareInstance,
		status.SharedInstanceMode,
		status.ConnectedToSharedInstance,
	)

	t.Log("waiting for nomadnetwork.node announces...")
	nodes := waitForNodes(stack.Handler(), 120*time.Second, 1)
	if len(nodes) == 0 {
		t.Fatalf(
			"no nodes discovered within timeout (online=%d outbound=%d config=%s)",
			len(allOnline),
			len(outbound),
			stack.ConfigPath(),
		)
	}
	t.Logf("discovered %d node(s)", len(nodes))

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	const maxNodes = 6
	const retries = 2
	success := 0
	ifaceHits := map[string]int{}

	for i, node := range nodes {
		if i >= maxNodes {
			break
		}
		var lastErr string
		var lastIface string
		var lastHops int
		ok := false
		for attempt := 0; attempt <= retries; attempt++ {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
				stack.Browser().RefreshPathForRetry(node.Hash)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			fetch := stack.Browser().Fetch(ctx, node.Hash, "/page/index.mu", nomadnet.RequestData{})
			cancel()
			lastErr = fetch.Error
			lastIface = fetch.Interface
			lastHops = fetch.Hops
			if fetch.Error == "" && len(fetch.Body) > 0 {
				success++
				ok = true
				ifaceHits[fetch.Interface]++
				t.Logf(
					"ok: %s (%s) hops=%d iface=%s %d bytes in %dms",
					node.Name,
					node.Hash,
					fetch.Hops,
					fetch.Interface,
					len(fetch.Body),
					fetch.DurationMs,
				)
				break
			}
			t.Logf(
				"attempt %d failed: %s (%s) hops=%d iface=%s err=%s",
				attempt+1,
				node.Name,
				node.Hash,
				fetch.Hops,
				fetch.Interface,
				fetch.Error,
			)
		}
		if !ok {
			t.Logf("unreachable after retries: %s (%s) hops=%d iface=%s: %s", node.Name, node.Hash, lastHops, lastIface, lastErr)
		}
	}

	t.Logf("interface hit counts: %v", ifaceHits)
	if success == 0 {
		t.Fatal("could not fetch any discovered node")
	}
}

// TestLiveWakePathReload maps the post-suspend reload path: after a successful
// fetch, invalidate the cached route (as PrepareForWake / soft-stale expiry
// would), then fetch again and require rediscovery to succeed.
func TestLiveWakePathReload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped with -short")
	}

	root := t.TempDir()
	paths.SetDataRoot(root)
	t.Cleanup(func() { paths.SetDataRoot("") })

	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	ensureInteropCommunityInterfaces(t, stack)
	if err := stack.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stack.Stop() }()

	if len(waitForOutboundOnlineInterfaces(stack, 90*time.Second, 1)) == 0 {
		t.Fatal("no online outbound interfaces")
	}
	nodes := waitForNodes(stack.Handler(), 120*time.Second, 1)
	if len(nodes) == 0 {
		t.Fatal("no nodes discovered")
	}

	browser := stack.Browser()
	var node nomadnet.Node
	var first nomadnet.FetchResult
	ok := false
	for _, candidate := range nodes {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		first = browser.Fetch(ctx, candidate.Hash, "/page/index.mu", nomadnet.RequestData{})
		cancel()
		if first.Error == "" && len(first.Body) > 0 {
			node = candidate
			ok = true
			break
		}
	}
	if !ok {
		t.Fatal("could not fetch any node before wake simulation")
	}
	t.Logf(
		"pre-wake fetch ok: %s hops=%d iface=%s %dms",
		node.Name, first.Hops, first.Interface, first.DurationMs,
	)

	wake := browser.PrepareForWake()
	t.Logf("PrepareForWake after fresh traffic: droppedLinks=%d expiredPaths=%d", wake.DroppedLinks, wake.ExpiredPaths)

	// Simulate soft-stale / post-suspend invalidation of this destination.
	browser.RefreshPathForRetry(node.Hash)
	wake2 := browser.PrepareForWake()
	t.Logf("PrepareForWake after forced expire: droppedLinks=%d expiredPaths=%d", wake2.DroppedLinks, wake2.ExpiredPaths)

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	second := browser.Fetch(ctx, node.Hash, "/page/index.mu", nomadnet.RequestData{})
	cancel()
	elapsed := time.Since(start)
	t.Logf(
		"post-wake fetch: err=%q hops=%d iface=%s bytes=%d %dms wall=%s",
		second.Error, second.Hops, second.Interface, len(second.Body), second.DurationMs, elapsed.Round(time.Millisecond),
	)
	if second.Error != "" || len(second.Body) == 0 {
		t.Fatalf("post-wake reload failed: %s", second.Error)
	}
}

func TestPickRandomCommunityInterfacesLiveQuality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped with -short")
	}
	result, err := rns.FetchCommunityInterfaces(nil)
	if err != nil {
		t.Fatal(err)
	}
	picked := rns.PickSeedableCommunityInterfaces(result.Items, rns.DefaultCommunityInterfaceCount)
	if len(picked) == 0 {
		t.Fatal("no seedable community interfaces")
	}
	t.Logf("picked %d/%d (fromBundle=%v)", len(picked), rns.DefaultCommunityInterfaceCount, result.FromBundle)
	seen := map[string]bool{}
	for _, item := range picked {
		t.Logf("  %s type=%s network=%s host=%s status=%s", item.Name, item.Type, item.Network, item.Host, item.Status)
		if item.Network == "i2p" || item.Network == "yggdrasil" {
			t.Fatalf("picked overlay interface %s (%s)", item.Name, item.Network)
		}
		if item.Status != "" && item.Status != "online" {
			t.Fatalf("picked non-online interface %s (%s)", item.Name, item.Status)
		}
		key := item.Host
		if key == "" {
			continue
		}
		if seen[key] {
			t.Fatalf("duplicate host selected: %s", key)
		}
		seen[key] = true
	}
	if len(picked) < 4 {
		t.Fatalf("expected at least 4 preferred clearnet picks, got %d", len(picked))
	}
}

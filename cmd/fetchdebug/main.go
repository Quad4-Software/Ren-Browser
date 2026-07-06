// SPDX-License-Identifier: MIT
//
// fetchdebug is a standalone reproduction tool for diagnosing Nomad Network
// page/file fetch failures against a real Reticulum interface, independent
// of the Wails desktop app. It reuses RenBrowser's own profile config (so it
// sees the same interfaces and identity the app uses) and prints every
// stage of Browser.FetchWithHooks as it happens, including resource
// transfer byte progress, so a hang or an "empty response" can be traced to
// the exact step (path discovery, link establishment, request send, or
// response wait) and the exact reticulum-go status it failed with.
//
// Usage:
//
//	go run ./cmd/fetchdebug -node <32-hex-char-node-hash> -page /page/index.mu -file /file/some/file.bin
//
// Pass -config to point at a different profile's Reticulum config file, and
// -wait to control how long to wait for the target node to announce itself
// before giving up.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"quad4/reticulum-go/pkg/debug"

	"renbrowser/internal/nomadnet"
	"renbrowser/internal/rns"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		configPath   = flag.String("config", "", "path to reticulum config file (default: RenBrowser's default profile)")
		nodeHash     = flag.String("node", "", "32 hex char destination hash of the node to fetch from (required)")
		pagePath     = flag.String("page", "", "optional /page/ path to fetch first, as a warm-up/connectivity check")
		filePath     = flag.String("file", "", "optional /file/ path to fetch, for reproducing download issues")
		waitAnnounce = flag.Duration("wait", 45*time.Second, "how long to wait for the node to announce itself before giving up")
		debugLevel   = flag.Int("debug-level", debug.DebugAll, "reticulum-go debug level (1=critical .. 7=all)")
	)
	flag.Parse()

	if *nodeHash == "" {
		fmt.Fprintln(os.Stderr, "error: -node is required")
		flag.Usage()
		return 2
	}
	if *pagePath == "" && *filePath == "" {
		fmt.Fprintln(os.Stderr, "error: at least one of -page or -file is required")
		flag.Usage()
		return 2
	}

	node := strings.ToLower(strings.TrimSpace(*nodeHash))

	stack, err := rns.NewStack(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build reticulum stack: %v\n", err)
		return 1
	}
	debug.SetDebugLevel(*debugLevel)
	debug.Init()

	logf("using config: %s", stack.ConfigPath())
	if err := stack.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start reticulum stack: %v\n", err)
		return 1
	}
	defer func() {
		_ = stack.Stop()
	}()

	for _, iface := range stack.ListInterfaces() {
		logf("interface: %s enabled=%v", iface.Name, iface.Enabled)
	}

	// Path resolution (routing) and identity resolution (the destination's
	// public keys, needed to establish an encrypted link) are separate in
	// RNS: a path can resolve via another node's cached routing table entry
	// without ever handing us a full announce. So don't hard-gate the fetch
	// on having already cached an announce ourselves -- that only tells us
	// whether a link can be *established*, not whether a route exists. Do a
	// short best-effort wait for a head start, then let Fetch's own stage
	// hooks report exactly which step (path vs. identity vs. link vs.
	// request vs. response) actually stalls or fails.
	awaitAnnounce(stack, node, min(*waitAnnounce, 20*time.Second))

	deadline := time.Now().Add(*waitAnnounce)
	ok := true
	if *pagePath != "" {
		if !fetchOneWithAnnounceRetry(stack, node, *pagePath, deadline) {
			ok = false
		}
	}
	if *filePath != "" {
		if !fetchOneWithAnnounceRetry(stack, node, *filePath, deadline) {
			ok = false
		}
	}

	if !ok {
		return 1
	}
	return 0
}

func awaitAnnounce(stack *rns.Stack, node string, timeout time.Duration) bool {
	if n, ok := stack.Handler().Get(node); ok {
		logf("node already known: hash=%s name=%q hops=%d", n.Hash, n.Name, n.Hops)
		return true
	}
	logf("waiting up to %s for %s to announce...", timeout, node)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if n, ok := stack.Handler().Get(node); ok {
			logf("announce received: hash=%s name=%q hops=%d", n.Hash, n.Name, n.Hops)
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

// fetchOneWithAnnounceRetry attempts the fetch immediately so a real path or
// request-layer failure shows up right away, and only falls back to waiting
// for an announce (then retrying once) if the failure was specifically the
// "identity not known yet" case.
func fetchOneWithAnnounceRetry(stack *rns.Stack, node, path string, deadline time.Time) bool {
	if fetchOne(stack, node, path) {
		return true
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return false
	}
	logf("retrying after waiting up to %s more for an announce from %s", remaining.Round(time.Second), node)
	if !awaitAnnounce(stack, node, remaining) {
		logf("still never saw an announce from %s", node)
		return false
	}
	return fetchOne(stack, node, path)
}

func fetchOne(stack *rns.Stack, node, path string) bool {
	logf("---")
	logf("fetching %s:%s", node, path)
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	var lastPct = -1
	hooks := &nomadnet.FetchHooks{
		OnStage: func(stage, detail string) {
			logf("[%s] %s: %s", stage, path, detail)
		},
		OnProgress: func(p nomadnet.FetchProgress) {
			if p.Total <= 0 {
				logf("[progress] %s: %d bytes received (total unknown)", path, p.Received)
				return
			}
			pct := int(float64(p.Received) / float64(p.Total) * 100)
			if pct != lastPct {
				logf("[progress] %s: %d/%d bytes (%d%%)", path, p.Received, p.Total, pct)
				lastPct = pct
			}
		},
	}

	result := stack.Browser().FetchWithHooks(ctx, node, path, nomadnet.RequestData{}, hooks)
	elapsed := time.Since(start)

	if result.Error != "" {
		logf("FAILED after %s: %s", elapsed.Round(time.Millisecond), result.Error)
		return false
	}
	logf(
		"OK after %s: %d bytes, contentType=%s, fileName=%q, hops=%d, interface=%s",
		elapsed.Round(time.Millisecond), len(result.Body), result.ContentType, result.FileName, result.Hops, result.Interface,
	)
	return true
}

func logf(format string, args ...any) {
	fmt.Printf("%s %s\n", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, args...))
}

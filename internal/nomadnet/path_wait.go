// SPDX-License-Identifier: MIT
package nomadnet

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	rlink "quad4/reticulum-go/pkg/link"
	"quad4/reticulum-go/pkg/transport"
)

var errInvalidPathDestination = errors.New("invalid destination")

const (
	pathPollInterval = 40 * time.Millisecond
	pathWaitDefault  = 45 * time.Second
)

// Soft reuse limits for mobile suspend/resume. Hard TTL in reticulum-go is
// PathRequestTTL (300s) and link keepalives are minutes-scale; after phone
// sleep sockets die while caches still look valid and reloads hang.
// Vars so unit tests can shrink the windows without waiting.
var (
	pathSoftMaxAge   = 90 * time.Second
	linkReuseMaxIdle = 90 * time.Second
)

// waitPath resolves a transport route before link establishment.
//
// Soft-stale cached routes (older than pathSoftMaxAge) are expired and
// rediscovered instead of reused. While waiting, NudgePathRequest is sent
// every PathRequestMI. On timeout the cached entry is expired so the next
// attempt cannot reuse a dead route.
func waitPath(ctx context.Context, tr *transport.Transport, destHash []byte, total time.Duration) error {
	if tr == nil || len(destHash) != 16 {
		return errInvalidPathDestination
	}
	if total <= 0 {
		total = pathWaitDefault
	}

	expireSoftStalePath(tr, destHash)

	switch tr.PrepareFreshPathRequest(destHash) {
	case transport.PrepareFreshInvalidDestination:
		return errInvalidPathDestination
	case transport.PrepareFreshReusedValidPath:
		return nil
	}
	if tr.HasPath(destHash) {
		return nil
	}

	deadline := time.Now().Add(total)
	nextNudge := time.Now().Add(transport.PathRequestMI)

	for {
		if tr.HasPath(destHash) {
			return nil
		}

		now := time.Now()
		if !now.Before(deadline) {
			tr.ExpirePath(destHash)
			return context.DeadlineExceeded
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !now.Before(nextNudge) {
			_ = tr.NudgePathRequest(destHash)
			nextNudge = now.Add(transport.PathRequestMI)
		}

		sleep := pathPollInterval
		if remaining := time.Until(deadline); sleep > remaining {
			sleep = remaining
		}
		if untilNudge := time.Until(nextNudge); untilNudge > 0 && untilNudge < sleep {
			sleep = untilNudge
		}

		timer := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func expireSoftStalePath(tr *transport.Transport, destHash []byte) {
	if tr == nil || len(destHash) != 16 {
		return
	}
	age := pathAge(tr, destHash)
	if age < 0 {
		return
	}
	// GetPathTable timestamps are Unix seconds, so age is coarse. Treat
	// maxAge <= 0 as "always expire" for tests; otherwise expire at/after maxAge.
	if pathSoftMaxAge <= 0 || age >= pathSoftMaxAge {
		tr.ExpirePath(destHash)
	}
}

// pathAge returns how old the cached path is, or -1 if missing.
// GetPathTable hashes are truncated (packet.TruncatedHashLength/8 bytes), so
// matching uses a prefix of destHash.
func pathAge(tr *transport.Transport, destHash []byte) time.Duration {
	if tr == nil || len(destHash) != 16 || !tr.HasPath(destHash) {
		return -1
	}
	for _, entry := range tr.GetPathTable(nil) {
		if len(entry.Hash) == 0 || len(entry.Hash) > len(destHash) {
			continue
		}
		match := true
		for i := 0; i < len(entry.Hash); i++ {
			if entry.Hash[i] != destHash[i] {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		if entry.Timestamp <= 0 {
			return 0
		}
		updated := time.Unix(int64(entry.Timestamp), 0)
		return time.Since(updated)
	}
	// HasPath but no table row (or truncated collision miss): treat as fresh.
	return 0
}

func pathWaitError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "path discovery timed out"
	}
	return fmt.Sprintf("%v", err)
}

func transportHops(tr *transport.Transport, destHash []byte, handler *AnnounceHandler, nodeHash string) int {
	if tr != nil {
		if hops := tr.HopsTo(destHash); hops < transport.PathfinderM {
			return int(hops)
		}
	}
	if handler != nil {
		if node, ok := handler.Get(nodeHash); ok {
			return int(node.Hops)
		}
	}
	return -1
}

// ShouldRefreshPath reports whether a fetch failure is likely caused by a
// stale or dead transport route and should force rediscovery on retry.
func ShouldRefreshPath(errMsg string) bool {
	msg := strings.ToLower(strings.TrimSpace(errMsg))
	if msg == "" {
		return false
	}
	switch {
	case strings.Contains(msg, "no path"),
		strings.Contains(msg, "path discovery"),
		strings.Contains(msg, "link establish"),
		strings.Contains(msg, "link timeout"),
		strings.Contains(msg, "timed out"),
		strings.Contains(msg, "timeout"),
		strings.Contains(msg, "unresponsive"),
		strings.Contains(msg, "connection lost"),
		strings.Contains(msg, "connection failed"),
		strings.Contains(msg, "not discovered"):
		return true
	default:
		return false
	}
}

// RefreshPathForRetry drops a cached route and any active browser link so the
// next Fetch attempt rediscovers a path (possibly via a different interface).
func (b *Browser) RefreshPathForRetry(nodeHash string) {
	if b == nil || b.tr == nil {
		return
	}
	destHash, err := decodeNodeHash(nodeHash)
	if err != nil {
		return
	}
	b.dropCachedLink(destHash)
	b.tr.ExpirePath(destHash)
	_ = b.tr.PrepareFreshPathRequest(destHash)
}

// PrepareForWake drops idle cached links and soft-stale paths after the app
// returns from background/suspend. Call before the next page reload so Fetch
// does not reuse zombie StatusActive links or dead next-hops.
func (b *Browser) PrepareForWake() WakePrepResult {
	res := WakePrepResult{}
	if b == nil {
		return res
	}
	// Snapshot destinations before dropping links so soft-stale path expiry
	// still covers destinations that only lived in the link cache.
	dests := b.knownPathDestinations()
	res.DroppedLinks = b.dropIdleCachedLinks(linkReuseMaxIdle)
	if b.tr != nil {
		res.ExpiredPaths = expireSoftStaleDestinations(b.tr, dests, pathSoftMaxAge)
	}
	return res
}

// WakePrepResult summarizes what PrepareForWake invalidated.
type WakePrepResult struct {
	DroppedLinks int `json:"droppedLinks"`
	ExpiredPaths int `json:"expiredPaths"`
}

func (b *Browser) knownPathDestinations() [][]byte {
	if b == nil {
		return nil
	}
	seen := make(map[string]struct{})
	var out [][]byte

	add := func(dest []byte) {
		if len(dest) != 16 {
			return
		}
		key := hex.EncodeToString(dest)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, append([]byte(nil), dest...))
	}

	b.mu.Lock()
	for key := range b.links {
		if dest, err := hex.DecodeString(key); err == nil {
			add(dest)
		}
	}
	b.mu.Unlock()

	if b.handler != nil {
		for _, node := range b.handler.List() {
			if dest, err := decodeNodeHash(node.Hash); err == nil {
				add(dest)
			}
		}
	}
	return out
}

func expireSoftStaleDestinations(tr *transport.Transport, dests [][]byte, maxAge time.Duration) int {
	if tr == nil {
		return 0
	}
	expired := 0
	for _, dest := range dests {
		if len(dest) != 16 || !tr.HasPath(dest) {
			continue
		}
		age := pathAge(tr, dest)
		if age < 0 {
			continue
		}
		if maxAge > 0 && age < maxAge {
			continue
		}
		tr.ExpirePath(dest)
		expired++
	}
	return expired
}

func decodeNodeHash(nodeHash string) ([]byte, error) {
	destHash, err := hex.DecodeString(normalizeHash(nodeHash))
	if err != nil || len(destHash) != 16 {
		return nil, errInvalidPathDestination
	}
	return destHash, nil
}

func (b *Browser) dropCachedLink(destHash []byte) {
	if b == nil || len(destHash) != 16 {
		return
	}
	key := hex.EncodeToString(destHash)
	b.mu.Lock()
	defer b.mu.Unlock()
	if cached := b.links[key]; cached != nil {
		cached.Teardown()
		delete(b.links, key)
	}
}

func (b *Browser) dropIdleCachedLinks(maxIdle time.Duration) int {
	if b == nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	dropped := 0
	for key, lnk := range b.links {
		if lnk == nil {
			delete(b.links, key)
			dropped++
			continue
		}
		if lnk.GetStatus() != rlink.StatusActive || linkIdleDuration(lnk) > maxIdle {
			lnk.Teardown()
			delete(b.links, key)
			dropped++
		}
	}
	return dropped
}

func linkIdleDuration(lnk *rlink.Link) time.Duration {
	if lnk == nil {
		return 0
	}
	sec := lnk.InactiveFor()
	if sec <= 0 {
		return 0
	}
	return time.Duration(sec * float64(time.Second))
}

// SPDX-License-Identifier: MIT
package nomadnet

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"quad4/reticulum-go/pkg/destination"
	"quad4/reticulum-go/pkg/identity"
	rlink "quad4/reticulum-go/pkg/link"
	"quad4/reticulum-go/pkg/transport"

	"renbrowser/internal/limits"
)

const (
	defaultRequestTimeout = 20 * time.Second
	defaultReceiptTimeout = 25 * time.Second
	fileRequestTimeout    = 280 * time.Second
	fileReceiptTimeout    = 285 * time.Second
)

// requestTimeouts picks how long to wait for a response before giving up.
// /file/ responses are delivered as RNS resource transfers that can span
// many packets and take far longer than a single-packet page response,
// especially over slow interfaces; using the page-sized timeout for both
// caused large file downloads to be aborted as "empty response" well
// before the transfer had a chance to finish.
func requestTimeouts(path string) (request, receipt time.Duration) {
	if strings.HasPrefix(path, "/file/") {
		return fileRequestTimeout, fileReceiptTimeout
	}
	return defaultRequestTimeout, defaultReceiptTimeout
}

// FetchProgress reports how much of a /file/ resource transfer has arrived
// so far. Total is 0 until the resource advertisement carrying the transfer
// size has been received.
type FetchProgress struct {
	Received int64
	Total    int64
}

// FetchHooks lets callers observe the stages of a Fetch call without
// changing its return type. Both fields are optional; nil hooks (or a nil
// *FetchHooks) disable all observability with no extra cost. OnStage fires
// once per named stage transition (e.g. "path", "link", "request",
// "waiting"). OnProgress fires whenever the receiving byte count changes
// while waiting on a resource-backed response (large /file/ downloads).
type FetchHooks struct {
	OnStage    func(stage, detail string)
	OnProgress func(FetchProgress)
}

func (h *FetchHooks) stage(stage, detail string) {
	if h != nil && h.OnStage != nil {
		h.OnStage(stage, detail)
	}
}

func (h *FetchHooks) progress(p FetchProgress) {
	if h != nil && h.OnProgress != nil {
		h.OnProgress(p)
	}
}

type FetchResult struct {
	NodeHash    string `json:"nodeHash"`
	Path        string `json:"path"`
	Body        []byte `json:"body"`
	ContentType string `json:"contentType"`
	FileName    string `json:"fileName,omitempty"`
	DurationMs  int64  `json:"durationMs"`
	Hops        int    `json:"hops"`
	Interface   string `json:"interface,omitempty"`
	Error       string `json:"error,omitempty"`
}

type Browser struct {
	mu           sync.Mutex
	tr           *transport.Transport
	handler      *AnnounceHandler
	links        map[string]*rlink.Link
	establishing map[string]*linkEstablishWait
}

type linkEstablishWait struct {
	done chan struct{}
	link *rlink.Link
	err  error
}

func NewBrowser(tr *transport.Transport, handler *AnnounceHandler) *Browser {
	return &Browser{
		tr:      tr,
		handler: handler,
		links:   make(map[string]*rlink.Link),
	}
}

func (b *Browser) Fetch(ctx context.Context, nodeHash string, path string, req RequestData) FetchResult {
	return b.FetchWithHooks(ctx, nodeHash, path, req, nil)
}

// FetchWithHooks is Fetch with optional stage/progress observability. See
// FetchHooks for details; pass nil for the same behavior as Fetch.
func (b *Browser) FetchWithHooks(ctx context.Context, nodeHash string, path string, req RequestData, hooks *FetchHooks) FetchResult {
	start := time.Now()
	res := FetchResult{
		NodeHash: normalizeHash(nodeHash),
		Path:     normalizePath(path),
	}

	destHash, err := hex.DecodeString(res.NodeHash)
	if err != nil || len(destHash) != 16 {
		res.Error = "invalid node hash"
		return res
	}

	// Reusing an already-established link means the route to this node is
	// known good right now, so skip path discovery entirely: the transport's
	// path table entry can otherwise expire (PathRequestTTL) well before a
	// quiet-but-active link goes stale, which would needlessly force a fresh
	// path request (and its poll/nudge wait) on every subsequent page load
	// to a node the user is already connected to.
	lnk := b.cachedLink(destHash)
	if lnk != nil {
		hooks.stage("link", "reusing cached active link")
	} else {
		hooks.stage("path", "waiting for path to node")
		if err := waitPath(ctx, b.tr, destHash, pathWaitDefault); err != nil {
			res.Hops = transportHops(b.tr, destHash, b.handler, res.NodeHash)
			res.Error = fmt.Sprintf("no path to node: %s", pathWaitError(err))
			res.DurationMs = time.Since(start).Milliseconds()
			hooks.stage("failed", res.Error)
			return res
		}
	}
	res.Hops = transportHops(b.tr, destHash, b.handler, res.NodeHash)
	res.Interface = b.tr.NextHopInterface(destHash)
	hooks.stage("path", fmt.Sprintf("path known, hops=%d interface=%s", res.Hops, res.Interface))

	if lnk == nil {
		remoteID, ok := b.handler.Identity(res.NodeHash)
		if !ok || remoteID == nil {
			res.Error = "node not discovered yet"
			res.DurationMs = time.Since(start).Milliseconds()
			hooks.stage("failed", res.Error)
			return res
		}

		hooks.stage("link", "establishing link")
		lnk, err = b.linkFor(ctx, destHash, remoteID)
		if err != nil {
			res.Error = err.Error()
			res.DurationMs = time.Since(start).Milliseconds()
			hooks.stage("failed", res.Error)
			return res
		}
		hooks.stage("link", "link established")
	}

	requestTimeout, receiptTimeout := requestTimeouts(res.Path)
	hooks.stage("request", fmt.Sprintf("sending request path=%s requestTimeout=%s", res.Path, requestTimeout))
	receipt, err := lnk.Request(res.Path, buildRequestData(req), requestTimeout)
	if err != nil {
		res.Error = err.Error()
		res.DurationMs = time.Since(start).Milliseconds()
		hooks.stage("failed", res.Error)
		return res
	}
	if iface := lnk.LinkedNetworkInterface(); iface != nil {
		res.Interface = iface.GetName()
	}

	hooks.stage("waiting", fmt.Sprintf("waiting for response, receiptTimeout=%s", receiptTimeout))
	maxBytes := limits.MaxFetchBytes(res.Path)
	body, metadata, err := waitReceipt(ctx, receipt, receiptTimeout, maxBytes, hooks)
	if err != nil {
		res.Error = err.Error()
		res.DurationMs = time.Since(start).Milliseconds()
		hooks.stage("failed", res.Error)
		return res
	}

	res.Body = body
	res.FileName = fileNameFromMetadata(metadata)
	res.ContentType = DetectContentType(res.Path, body)
	res.DurationMs = time.Since(start).Milliseconds()
	hooks.stage("done", fmt.Sprintf("received %d bytes in %dms", len(body), res.DurationMs))
	return res
}

// fileNameFromMetadata extracts the file name Nomad Network attaches to
// /file/ resource responses (serve_file returns [handle, {"name": ...}] on
// the Python side). Returns "" if no usable name is present.
func fileNameFromMetadata(metadata map[string]any) string {
	if metadata == nil {
		return ""
	}
	switch name := metadata["name"].(type) {
	case []byte:
		return string(name)
	case string:
		return name
	default:
		return ""
	}
}

// cachedLink returns the currently active link for destHash, if any, without
// establishing a new one or touching path discovery. Idle links (no traffic
// beyond linkReuseMaxIdle) are torn down so post-suspend reloads cannot reuse
// zombie StatusActive entries whose keepalives never ran.
func (b *Browser) cachedLink(destHash []byte) *rlink.Link {
	key := hex.EncodeToString(destHash)
	b.mu.Lock()
	defer b.mu.Unlock()
	cached := b.links[key]
	if cached == nil {
		delete(b.links, key)
		return nil
	}
	if cached.GetStatus() != rlink.StatusActive || linkIdleDuration(cached) > linkReuseMaxIdle {
		cached.Teardown()
		delete(b.links, key)
		return nil
	}
	return cached
}

func (b *Browser) linkFor(ctx context.Context, destHash []byte, remoteID *identity.Identity) (*rlink.Link, error) {
	key := hex.EncodeToString(destHash)

	b.mu.Lock()
	if cached := b.links[key]; cached != nil && cached.GetStatus() == rlink.StatusActive && linkIdleDuration(cached) <= linkReuseMaxIdle {
		b.mu.Unlock()
		return cached, nil
	}
	if cached := b.links[key]; cached != nil {
		cached.Teardown()
		delete(b.links, key)
	}
	waiter, already := b.establishing[key]
	if already {
		b.mu.Unlock()
		select {
		case <-waiter.done:
			if waiter.err != nil {
				return nil, waiter.err
			}
			return waiter.link, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	waiter = &linkEstablishWait{done: make(chan struct{})}
	if b.establishing == nil {
		b.establishing = make(map[string]*linkEstablishWait)
	}
	b.establishing[key] = waiter
	b.mu.Unlock()

	lnk, err := b.establishLink(ctx, destHash, remoteID)

	waiter.link = lnk
	waiter.err = err
	close(waiter.done)

	b.mu.Lock()
	delete(b.establishing, key)
	if err == nil {
		b.links[key] = lnk
	}
	b.mu.Unlock()
	return lnk, err
}

func (b *Browser) establishLink(ctx context.Context, destHash []byte, remoteID *identity.Identity) (*rlink.Link, error) {
	destOut, err := destination.FromHash(destHash, remoteID, destination.Single, b.tr)
	if err != nil {
		return nil, err
	}

	established := make(chan struct{}, 1)
	lnk := rlink.NewLink(destOut, b.tr, nil, func(_ *rlink.Link) {
		select {
		case established <- struct{}{}:
		default:
		}
	}, nil)

	if err := lnk.Establish(); err != nil {
		b.tr.ExpirePath(destHash)
		return nil, fmt.Errorf("link establish: %w", err)
	}

	select {
	case <-established:
	case <-ctx.Done():
		lnk.Teardown()
		b.tr.ExpirePath(destHash)
		return nil, fmt.Errorf("no path to node: %s", pathWaitError(ctx.Err()))
	case <-time.After(45 * time.Second):
		lnk.Teardown()
		b.tr.ExpirePath(destHash)
		return nil, fmt.Errorf("link establish timeout")
	}

	lnk.Start()
	return lnk, nil
}

func receiptStatusLabel(status byte) string {
	switch status {
	case rlink.StatusPending:
		return "pending"
	case rlink.StatusActive:
		return "active"
	case rlink.StatusFailed:
		return "failed"
	default:
		return fmt.Sprintf("unknown(%d)", status)
	}
}

func (b *Browser) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for key, lnk := range b.links {
		if lnk != nil {
			lnk.Teardown()
		}
		delete(b.links, key)
	}
	b.establishing = nil
}

func waitReceipt(ctx context.Context, receipt *rlink.RequestReceipt, timeout time.Duration, maxBytes int, hooks *FetchHooks) ([]byte, map[string]any, error) {
	start := time.Now()
	deadline := start.Add(timeout)
	var lastReceived int64 = -1
	var lastLogAt time.Time
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
		}
		if received, total := receipt.Progress(); received != lastReceived || time.Since(lastLogAt) >= 2*time.Second {
			if maxBytes > 0 && (total > int64(maxBytes) || received > int64(maxBytes)) {
				limit := maxBytes
				if total > int64(maxBytes) {
					return nil, nil, fmt.Errorf("response too large: advertised %d bytes (limit %d)", total, limit)
				}
				return nil, nil, fmt.Errorf("response too large: received %d bytes (limit %d)", received, limit)
			}
			if received != lastReceived {
				hooks.progress(FetchProgress{Received: received, Total: total})
			}
			if time.Since(lastLogAt) >= 2*time.Second {
				hooks.stage("waiting", fmt.Sprintf(
					"elapsed=%s status=%s received=%d total=%d",
					time.Since(start).Round(time.Millisecond), receiptStatusLabel(receipt.GetStatus()), received, total,
				))
				lastLogAt = time.Now()
			}
			lastReceived = received
		}
		if receipt.Concluded() {
			resp := receipt.GetResponse()
			if len(resp) == 0 {
				received, total := receipt.Progress()
				return nil, nil, fmt.Errorf(
					"empty response (status=%s received=%d total=%d elapsed=%s)",
					receiptStatusLabel(receipt.GetStatus()), received, total, time.Since(start).Round(time.Millisecond),
				)
			}
			_, total := receipt.Progress()
			if err := CheckResponseSize(resp, total, maxBytes); err != nil {
				return nil, nil, err
			}
			return resp, receipt.GetMetadata(), nil
		}
		time.Sleep(40 * time.Millisecond)
	}
	received, total := receipt.Progress()
	return nil, nil, fmt.Errorf(
		"node response timed out (received=%d total=%d after %s)",
		received, total, time.Since(start).Round(time.Millisecond),
	)
}

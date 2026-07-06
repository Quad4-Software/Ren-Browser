// SPDX-License-Identifier: MIT
package nomadnet

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"quad4/reticulum-go/pkg/destination"
	"quad4/reticulum-go/pkg/identity"
	rlink "quad4/reticulum-go/pkg/link"
	"quad4/reticulum-go/pkg/transport"
)

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

	if err := waitPath(ctx, b.tr, destHash, pathWaitDefault); err != nil {
		res.Hops = transportHops(b.tr, destHash, b.handler, res.NodeHash)
		res.Error = fmt.Sprintf("no path to node: %s", pathWaitError(err))
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}
	res.Hops = transportHops(b.tr, destHash, b.handler, res.NodeHash)
	res.Interface = b.tr.NextHopInterface(destHash)

	remoteID, ok := b.handler.Identity(res.NodeHash)
	if !ok || remoteID == nil {
		res.Error = "node not discovered yet"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	lnk, err := b.linkFor(ctx, destHash, remoteID)
	if err != nil {
		res.Error = err.Error()
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	receipt, err := lnk.Request(res.Path, buildRequestData(req), 20*time.Second)
	if err != nil {
		res.Error = err.Error()
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}
	if iface := lnk.LinkedNetworkInterface(); iface != nil {
		res.Interface = iface.GetName()
	}

	body, metadata, err := waitReceipt(ctx, receipt, 25*time.Second)
	if err != nil {
		res.Error = err.Error()
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	res.Body = body
	res.FileName = fileNameFromMetadata(metadata)
	res.ContentType = DetectContentType(res.Path, body)
	res.DurationMs = time.Since(start).Milliseconds()
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

func (b *Browser) linkFor(ctx context.Context, destHash []byte, remoteID *identity.Identity) (*rlink.Link, error) {
	key := hex.EncodeToString(destHash)

	b.mu.Lock()
	if cached := b.links[key]; cached != nil && cached.GetStatus() == rlink.StatusActive {
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
		return nil, fmt.Errorf("link establish: %w", err)
	}

	select {
	case <-established:
	case <-ctx.Done():
		lnk.Teardown()
		return nil, fmt.Errorf("no path to node: %s", pathWaitError(ctx.Err()))
	case <-time.After(45 * time.Second):
		lnk.Teardown()
		return nil, fmt.Errorf("link establish timeout")
	}

	lnk.Start()
	return lnk, nil
}

func waitReceipt(ctx context.Context, receipt *rlink.RequestReceipt, timeout time.Duration) ([]byte, map[string]any, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
		}
		if receipt.Concluded() {
			resp := receipt.GetResponse()
			if len(resp) == 0 {
				return nil, nil, fmt.Errorf("empty response")
			}
			return resp, receipt.GetMetadata(), nil
		}
		time.Sleep(80 * time.Millisecond)
	}
	return nil, nil, fmt.Errorf("node response timed out")
}

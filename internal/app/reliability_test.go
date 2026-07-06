// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"testing"
	"time"

	"renbrowser/internal/apperrors"
)

func TestApplyPageErrorPayloadTooLarge(t *testing.T) {
	resp := PageResponse{URL: "deadbeef:/page/index.mu"}
	applyPageError(&resp, "response too large: received 100 bytes (limit 8)", nil)
	if resp.ErrorKind != string(apperrors.KindPayloadTooLarge) {
		t.Fatalf("kind=%q", resp.ErrorKind)
	}
	if resp.Error == "" {
		t.Fatal("expected error detail")
	}
}

func TestApplyPageErrorShuttingDown(t *testing.T) {
	resp := PageResponse{}
	applyPageError(&resp, "application shutting down", nil)
	if resp.ErrorKind == "" {
		t.Fatal("expected classified error kind")
	}
}

func TestDownloadManagerCancelAllInvokesBoundCancel(t *testing.T) {
	m := newDownloadManager()
	id := m.start("nomad://node/file/a.bin", "a.bin")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.bindCancel(id, cancel)

	m.cancelAll()

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("download cancel func was not invoked")
	}

	items := m.list()
	if len(items) != 1 || items[0].Status != DownloadStatusCanceled {
		t.Fatalf("items=%#v", items)
	}
}

func TestDownloadManagerCancelAllSkipsFinished(t *testing.T) {
	m := newDownloadManager()
	id := m.start("nomad://node/file/a.bin", "a.bin")
	m.complete(id, "/tmp/a.bin", 10)

	m.cancelAll()

	items := m.list()
	if len(items) != 1 || items[0].Status != DownloadStatusCompleted {
		t.Fatalf("items=%#v", items)
	}
}

// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"path/filepath"
	"testing"

	"renbrowser/internal/rns"
)

func TestReconcilePendingJobAsInterrupted(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "profile.db")
	cfgPath := filepath.Join(dir, "config")

	svc1, err := NewBrowserServiceWithOptions(nil, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	svc1.addPendingDownloadJob("deadbeef:/file/guide.zip", "guide.zip")
	if err := svc1.Store().Close(); err != nil {
		t.Fatal(err)
	}

	stack, err := rns.NewStack(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stack.Stop() })

	svc2, err := NewBrowserServiceWithOptions(stack, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = svc2.Store().Close() })

	if len(svc2.loadPendingDownloadJobs()) != 0 {
		t.Fatalf("pending jobs=%#v", svc2.loadPendingDownloadJobs())
	}
	items := svc2.downloads.list()
	if len(items) != 1 {
		t.Fatalf("items=%#v", items)
	}
	if items[0].Status != DownloadStatusInterrupted {
		t.Fatalf("status=%q", items[0].Status)
	}
	if items[0].URL != "deadbeef:/file/guide.zip" {
		t.Fatalf("url=%q", items[0].URL)
	}
	if svc2.downloads.runningCount() != 0 {
		t.Fatal("reconcile should not auto-start downloads")
	}
}

func TestDownloadRecoveryPersistRoundTrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "profile.db")
	cfgPath := filepath.Join(dir, "config")

	svc1, err := NewBrowserServiceWithOptions(nil, nil, ServiceOptions{ProfilePath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	id := svc1.downloads.start("deadbeef:/file/a.bin", "a.bin")
	svc1.downloads.fail(id, "node response timed out")
	svc1.persistDownloadRecovery(svc1.downloads.list())
	if err := svc1.Store().Close(); err != nil {
		t.Fatal(err)
	}

	svc2 := reopenTestBrowserService(t, dbPath, cfgPath)
	items := svc2.downloads.list()
	if len(items) != 1 {
		t.Fatalf("items=%#v", items)
	}
	if items[0].Status != DownloadStatusFailed {
		t.Fatalf("status=%q", items[0].Status)
	}
	if items[0].Error != "node response timed out" {
		t.Fatalf("error=%q", items[0].Error)
	}
}

func TestShutdownInFlightMarksInterrupted(t *testing.T) {
	m := newDownloadManager()
	id := m.start("deadbeef:/file/a.bin", "a.bin")
	m.reportProgress(id, 1024, 4096)
	m.shutdownInFlight(downloadInterruptedText)
	m.markFailedUnlessCanceled(id, context.Canceled)

	items := m.list()
	if len(items) != 1 {
		t.Fatalf("items=%#v", items)
	}
	if items[0].Status != DownloadStatusInterrupted {
		t.Fatalf("status=%q", items[0].Status)
	}
	if items[0].Received != 1024 {
		t.Fatalf("received=%d", items[0].Received)
	}
}

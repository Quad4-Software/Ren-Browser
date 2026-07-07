// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRetriableDownloadError(t *testing.T) {
	if retriableDownloadError(nil) {
		t.Fatal("nil should not be retriable")
	}
	if retriableDownloadError(context.Canceled) {
		t.Fatal("user cancel should not be retriable")
	}
	if !retriableDownloadError(context.DeadlineExceeded) {
		t.Fatal("deadline exceeded should be retriable")
	}
	if !retriableDownloadError(errors.New("node response timed out")) {
		t.Fatal("timeout message should be retriable")
	}
	if retriableDownloadError(errors.New("permission denied")) {
		t.Fatal("permission denied should not be retriable")
	}
}

func TestPendingDownloadJobsPersist(t *testing.T) {
	svc := newTestBrowserService(t)
	svc.addPendingDownloadJob("deadbeef:/file/guide.zip", "guide.zip")
	jobs := svc.loadPendingDownloadJobs()
	if len(jobs) != 1 || jobs[0].URL != "deadbeef:/file/guide.zip" {
		t.Fatalf("jobs=%#v", jobs)
	}
	svc.removePendingDownloadJob("deadbeef:/file/guide.zip")
	if len(svc.loadPendingDownloadJobs()) != 0 {
		t.Fatal("expected pending jobs cleared")
	}
}

func TestCanceledDownloadStaysUntilDismissed(t *testing.T) {
	m := newDownloadManager()
	id := m.start("deadbeef:/file/a.bin", "a.bin")
	if !m.cancel(id) {
		t.Fatal("expected cancel to succeed")
	}
	items := m.list()
	if len(items) != 1 || items[0].Status != DownloadStatusCanceled {
		t.Fatalf("items=%#v", items)
	}
	m.dismiss(id)
	if len(m.list()) != 0 {
		t.Fatal("expected dismissed canceled download to be removed")
	}
}

func TestDownloadManagerCancelDoesNotMarkFailed(t *testing.T) {
	m := newDownloadManager()
	id := m.start("nomad://node/file/a.bin", "a.bin")
	m.cancel(id)
	m.markFailedUnlessCanceled(id, context.Canceled)
	items := m.list()
	if len(items) != 1 || items[0].Status != DownloadStatusCanceled {
		t.Fatalf("items=%#v", items)
	}
	if items[0].Error != "" {
		t.Fatalf("error=%q", items[0].Error)
	}
}

func TestRetryDownloadRequiresFailedState(t *testing.T) {
	svc := newTestBrowserService(t)
	id := svc.downloads.start("deadbeef:/file/a.bin", "a.bin")
	if svc.RetryDownload(id).OK {
		t.Fatal("pending download should not retry")
	}
	svc.downloads.complete(id, "/tmp/a.bin", 1)
	if svc.RetryDownload(id).OK {
		t.Fatal("completed download should not retry")
	}
}

func TestRetryDownloadAcceptsCanceledStatus(t *testing.T) {
	svc := newTestBrowserService(t)
	id := svc.downloads.start("deadbeef:/file/a.bin", "a.bin")
	svc.downloads.cancel(id)

	item, ok := svc.downloads.findByID(id)
	if !ok {
		t.Fatal("missing download")
	}
	if item.Status != DownloadStatusCanceled {
		t.Fatalf("status=%q", item.Status)
	}
	url, name, err := svc.resolveRetryDownload(item)
	if err != nil {
		t.Fatal(err)
	}
	if url != "deadbeef:/file/a.bin" || name != "a.bin" {
		t.Fatalf("url=%q name=%q", url, name)
	}
}

func TestRetryDownloadMissingURL(t *testing.T) {
	svc := newTestBrowserService(t)
	svc.downloads.importRecovery(downloadRecoveryRecord{
		ID:     "dl-1",
		Name:   "a.bin",
		Status: DownloadStatusFailed,
		Error:  "timeout",
	})
	result := svc.RetryDownload("dl-1")
	if result.OK {
		t.Fatal("expected retry to fail without URL")
	}
	if !strings.Contains(result.Error, "URL") {
		t.Fatalf("error=%q", result.Error)
	}
}

func TestResolveDownloadFilenameFromStalePath(t *testing.T) {
	svc := newTestBrowserService(t)
	name := svc.resolveDownloadFilename(ActiveDownload{
		URL:  "deadbeef:/file/guide.zip",
		Path: "/stale/missing/guide.zip",
	})
	if name != "guide.zip" {
		t.Fatalf("name=%q", name)
	}
}

func TestFindExistingDownloadPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(home, "renbrowser-dl-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	svc := newTestBrowserService(t)
	svc.SetDownloadDir(dir)
	path := filepath.Join(dir, "guide.zip")
	if err := writeTestFile(path); err != nil {
		t.Fatal(err)
	}
	found := svc.findExistingDownloadPath("guide.zip")
	if found != path {
		t.Fatalf("found=%q want=%q", found, path)
	}
}

func TestTrackedDownloadFailureCanceled(t *testing.T) {
	m := newDownloadManager()
	id := m.start("deadbeef:/file/a.bin", "a.bin")
	m.cancel(id)
	err := trackedDownloadFailure(m, id, context.Canceled)
	if !errors.Is(err, ErrDownloadCanceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestRetryDownloadPreparesFreshPendingJob(t *testing.T) {
	svc := newTestBrowserService(t)
	id := svc.downloads.start("deadbeef:/file/guide.zip", "guide.zip")
	svc.downloads.fail(id, "timeout")
	svc.addPendingDownloadJob("deadbeef:/file/guide.zip", "guide.zip")

	item, ok := svc.downloads.findByID(id)
	if !ok {
		t.Fatal("missing download")
	}
	svc.downloads.dismiss(id)
	svc.removePendingDownloadJob(item.URL)

	if len(svc.loadPendingDownloadJobs()) != 0 {
		t.Fatalf("jobs=%#v", svc.loadPendingDownloadJobs())
	}
}

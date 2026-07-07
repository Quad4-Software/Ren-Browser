// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"errors"
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

func TestRetryDownloadRequiresFailedState(t *testing.T) {
	svc := newTestBrowserService(t)
	id := svc.downloads.start("deadbeef:/file/a.bin", "a.bin")
	if svc.RetryDownload(id) {
		t.Fatal("pending download should not retry")
	}
	svc.downloads.complete(id, "/tmp/a.bin", 1)
	if svc.RetryDownload(id) {
		t.Fatal("completed download should not retry")
	}
}

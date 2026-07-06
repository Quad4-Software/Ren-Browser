// SPDX-License-Identifier: MIT
package nomadnet

import "testing"

func TestRequestTimeoutsFilePaths(t *testing.T) {
	req, receipt := requestTimeouts("/file/music/song.mp3")
	if req != fileRequestTimeout || receipt != fileReceiptTimeout {
		t.Fatalf("file timeouts = %v/%v; want %v/%v", req, receipt, fileRequestTimeout, fileReceiptTimeout)
	}
	if receipt <= req {
		t.Fatalf("receipt timeout %v must exceed request timeout %v", receipt, req)
	}
}

func TestRequestTimeoutsPagePaths(t *testing.T) {
	req, receipt := requestTimeouts("/page/index.mu")
	if req != defaultRequestTimeout || receipt != defaultReceiptTimeout {
		t.Fatalf("page timeouts = %v/%v; want %v/%v", req, receipt, defaultRequestTimeout, defaultReceiptTimeout)
	}
	if receipt <= req {
		t.Fatalf("receipt timeout %v must exceed request timeout %v", receipt, req)
	}
}

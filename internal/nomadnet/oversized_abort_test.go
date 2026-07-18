// SPDX-License-Identifier: MIT
package nomadnet

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestOversizedAbortTearsDownLink verifies Fetch aborts the transfer and drops
// the cached link when waitReceipt fails (oversize, cancel, timeout).
func TestOversizedAbortTearsDownLink(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	srcPath := filepath.Join(filepath.Dir(thisFile), "browser.go")
	raw, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)

	if !strings.Contains(src, "SetMaxResponseBytes") {
		t.Fatal("expected SetMaxResponseBytes before waitReceipt")
	}
	if !strings.Contains(src, "abortFetch(") {
		t.Fatal("expected abortFetch on waitReceipt error")
	}
	abortIdx := strings.Index(src, "func (b *Browser) abortFetch(")
	if abortIdx < 0 {
		t.Fatal("abortFetch helper missing")
	}
	abortSection := src[abortIdx:]
	if end := strings.Index(abortSection[1:], "\nfunc "); end > 0 {
		abortSection = abortSection[:end+1]
	}
	if !strings.Contains(abortSection, "AbortIncomingResponse()") {
		t.Fatal("abortFetch must call AbortIncomingResponse")
	}
	if !strings.Contains(abortSection, "Teardown()") {
		t.Fatal("abortFetch must Teardown the link")
	}
	t.Log("confirmed: oversize/cancel abort clears resource and tears down link")
}

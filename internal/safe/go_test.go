// SPDX-License-Identifier: MIT
package safe_test

import (
	"sync"
	"testing"
	"time"

	"renbrowser/internal/safe"
)

func TestGoRunsWithoutPanic(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	safe.Go("test-ok", func() {
		defer wg.Done()
	})
	wg.Wait()
}

func TestGoRecoversPanic(t *testing.T) {
	done := make(chan struct{})
	safe.Go("test-panic", func() {
		close(done)
		panic("boom")
	})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("goroutine did not start")
	}
	time.Sleep(50 * time.Millisecond)
}

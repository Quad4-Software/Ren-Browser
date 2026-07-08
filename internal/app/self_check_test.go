// SPDX-License-Identifier: MIT
package app

import (
	"testing"
)

func TestSelfCheck(t *testing.T) {
	svc := newTestBrowserService(t)

	// Since stack is not started in newTestBrowserService, StackUp should be false
	res := svc.RunSelfCheck()

	// We can start the stack to test the full flow
	stack := svc.stack
	if stack != nil {
		if err := stack.Start(); err != nil {
			t.Fatalf("failed to start stack: %v", err)
		}
	}

	res = svc.RunSelfCheck()

	if !res.StackUp.Passed {
		t.Errorf("expected StackUp to pass after starting, got: %s", res.StackUp.Reason)
	}

	if !res.ConfigGood.Passed {
		t.Errorf("expected ConfigGood to pass, got: %s", res.ConfigGood.Reason)
	}

	if !res.DBGood.Passed {
		t.Errorf("expected DBGood to pass, got: %s", res.DBGood.Reason)
	}

	if !res.ReadWriteGood.Passed {
		t.Errorf("expected ReadWriteGood to pass, got: %s", res.ReadWriteGood.Reason)
	}

	if !res.DownloadsGood.Passed {
		t.Errorf("expected DownloadsGood to pass, got: %s", res.DownloadsGood.Reason)
	}

	if !res.AllPassed {
		t.Errorf("expected AllPassed to be true, got: %+v", res)
	}
}

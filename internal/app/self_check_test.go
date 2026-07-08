// SPDX-License-Identifier: MIT
package app

import (
	"testing"
	"time"

	"renbrowser/internal/nomadnet"
)

func TestSelfCheck(t *testing.T) {
	svc := newTestBrowserService(t)

	stack := svc.stack
	if stack != nil {
		if err := stack.Start(); err != nil {
			t.Fatalf("failed to start stack: %v", err)
		}
	}

	res := svc.RunSelfCheck()

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
	if !res.Interfaces.Passed {
		t.Errorf("expected Interfaces to pass, got: %s", res.Interfaces.Reason)
	}
	if res.MeshEnabled {
		t.Fatal("mesh self-check should be off by default")
	}
	if !res.Discovery.Passed || !res.PageFetch.Passed {
		t.Fatalf("skipped mesh checks should pass: discovery=%+v page=%+v", res.Discovery, res.PageFetch)
	}
	if !res.AllPassed {
		t.Errorf("expected AllPassed to be true, got: %+v", res)
	}
}

func TestSelfCheckMeshEnvHelpers(t *testing.T) {
	t.Setenv(selfCheckMeshEnvKey, "1")
	if !selfCheckMeshEnabled() {
		t.Fatal("expected mesh enabled")
	}
	t.Setenv(selfCheckMeshEnvKey, "0")
	if selfCheckMeshEnabled() {
		t.Fatal("expected mesh disabled")
	}

	t.Setenv(selfCheckMeshCountEnvKey, "5")
	if selfCheckMeshCount() != 5 {
		t.Fatalf("count = %d", selfCheckMeshCount())
	}
	t.Setenv(selfCheckMeshCountEnvKey, "99")
	if selfCheckMeshCount() != maxSelfCheckMeshCount {
		t.Fatalf("count capped = %d", selfCheckMeshCount())
	}

	t.Setenv(selfCheckMeshWaitEnvKey, "30")
	if selfCheckMeshWait() != 30*time.Second {
		t.Fatalf("wait = %s", selfCheckMeshWait())
	}
}

func TestSelfCheckPageURLFromHandler(t *testing.T) {
	svc := newTestBrowserService(t)
	handler := svc.stack.Handler()
	if handler == nil {
		t.Fatal("nil announce handler")
	}
	handler.SetOnAnnounce(nil)
	// Seed a fake node via the handler map by calling List after injecting through ReceivedAnnounce
	// is awkward without a real identity; exercise FormatURL selection helper with empty list.
	_, err := svc.selfCheckPageURL(svc.stack)
	if err == nil {
		t.Fatal("expected error with no discovered nodes")
	}

	url := nomadnet.FormatURL("abb3ebcd03cb2388a838e70c001291f9", "/page/index.mu")
	if url != "abb3ebcd03cb2388a838e70c001291f9:/page/index.mu" {
		t.Fatalf("url = %q", url)
	}
}

func TestCloneReticulumConfig(t *testing.T) {
	svc := newTestBrowserService(t)
	cfg := svc.stack.Config()
	if cfg == nil {
		t.Fatal("nil config")
	}
	cloned := cloneReticulumConfig(cfg)
	if cloned == cfg {
		t.Fatal("expected distinct config pointer")
	}
	if &cloned.Interfaces == &cfg.Interfaces {
		t.Fatal("expected distinct interfaces map")
	}
}

func TestCheckInterfacesNoEnabled(t *testing.T) {
	svc := newTestBrowserService(t)
	if err := svc.stack.Start(); err != nil {
		t.Fatal(err)
	}
	res := SelfCheckResult{AllPassed: true}
	svc.checkInterfaces(&res)
	if !res.Interfaces.Passed {
		t.Fatalf("expected pass with no enabled ifaces: %s", res.Interfaces.Reason)
	}
}

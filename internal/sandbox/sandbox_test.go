// SPDX-License-Identifier: MIT
package sandbox

import (
	"path/filepath"
	"runtime"
	"testing"

	"renbrowser/internal/config"
	"renbrowser/internal/paths"
)

func TestRequested_DisabledByFlag(t *testing.T) {
	if Requested(Options{NoLandlock: true}) {
		t.Fatal("expected not requested when --no-landlock")
	}
}

func TestRequested_DesktopNotAuto(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	if !KernelSupported() {
		t.Skip("landlock not supported")
	}
	if Requested(Options{ServerMode: false}) {
		t.Fatal("desktop should not auto-enable landlock")
	}
	if !Requested(Options{ServerMode: true}) {
		t.Fatal("server should auto-enable landlock when supported")
	}
}

func TestApply_DesktopSkipped(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	Apply(Options{ServerMode: false})
	st := CurrentStatus()
	if st.Enabled {
		t.Fatal("expected landlock disabled on desktop by default")
	}
	if st.Reason == "" {
		t.Fatal("expected desktop skip reason")
	}
}

func TestRequested_DisabledByEnv(t *testing.T) {
	t.Setenv("REN_BROWSER_LANDLOCK", "0")
	if Requested(Options{}) {
		t.Fatal("expected not requested when REN_BROWSER_LANDLOCK=0")
	}
	if !DisabledByEnv() {
		t.Fatal("expected disabled by env")
	}
}

func TestRequested_ForceEnv(t *testing.T) {
	t.Setenv("REN_BROWSER_LANDLOCK", "1")
	if runtime.GOOS != "linux" {
		if Requested(Options{}) {
			t.Fatal("expected not requested off Linux even with env=1")
		}
		return
	}
	if !Requested(Options{}) {
		t.Fatal("expected requested when REN_BROWSER_LANDLOCK=1 on Linux")
	}
	if !Requested(Options{ServerMode: false}) {
		t.Fatal("expected env override on desktop too")
	}
}

func TestApply_RecordsStatus(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Linux apply tested via TestLandlockFunctional")
	}
	Apply(Options{})
	st := CurrentStatus()
	if st.Type != "none" {
		t.Fatalf("type = %q, want none", st.Type)
	}
	if st.Enabled {
		t.Fatal("expected disabled off Linux")
	}
	if st.Reason == "" {
		t.Fatal("expected reason when unsupported")
	}
}

func TestOptionsFromRuntime(t *testing.T) {
	tmp := t.TempDir()
	paths.SetDataRoot(tmp)
	t.Cleanup(func() { paths.SetDataRoot("") })

	opts := OptionsFromRuntime(config.Runtime{
		ReticulumConfig: filepath.Join(tmp, "custom", "config"),
		AssetsDir:       filepath.Join(tmp, "assets"),
	})
	if opts.DataDir == "" || opts.ReticulumDir == "" {
		t.Fatalf("unexpected empty dirs: %+v", opts)
	}
	if opts.PluginsDir == "" {
		t.Fatal("expected plugins dir")
	}
}

// SPDX-License-Identifier: MIT
package runtimeenv

import (
	"runtime"
	"testing"
)

func TestDetect_AndroidFlag(t *testing.T) {
	info := Detect()
	if runtime.GOOS == "android" && !info.OnAndroid {
		t.Fatal("expected OnAndroid on android GOOS")
	}
	if runtime.GOOS != "android" && info.OnAndroid {
		t.Fatal("expected OnAndroid false off android")
	}
}

func TestInAppImage(t *testing.T) {
	t.Setenv("APPIMAGE", "")
	if InAppImage() {
		t.Fatal("expected false without APPIMAGE")
	}
	t.Setenv("APPIMAGE", "/tmp/Ren.AppImage")
	if !InAppImage() {
		t.Fatal("expected true with APPIMAGE")
	}
}

func TestInFlatpak_Env(t *testing.T) {
	t.Setenv("FLATPAK_ID", "")
	// May still be true if /.flatpak-info exists on the host; only assert env path.
	t.Setenv("FLATPAK_ID", "io.quad4.renbrowser")
	if !InFlatpak() {
		t.Fatal("expected true with FLATPAK_ID")
	}
}

func TestWebKitSandbox_DisabledByEnv(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux desktop only")
	}
	t.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "1")
	t.Setenv("FLATPAK_ID", "")
	t.Setenv("APPIMAGE", "")
	state, note := webkitSandboxStatus(Info{})
	if state != "disabled" {
		t.Fatalf("state=%q want disabled", state)
	}
	if note != "env" {
		t.Fatalf("note=%q want env", note)
	}
}

func TestWebKitSandbox_UnavailableInContainer(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux only")
	}
	t.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "")
	state, note := webkitSandboxStatus(Info{InContainer: true, ContainerRuntime: "docker"})
	if state != "unavailable" {
		t.Fatalf("state=%q want unavailable", state)
	}
	if note != "container" {
		t.Fatalf("note=%q want container", note)
	}
}

func TestDetectContainer_Env(t *testing.T) {
	t.Setenv("container", "podman")
	ok, name := detectContainer()
	if !ok || name != "podman" {
		t.Fatalf("got ok=%v name=%q", ok, name)
	}
}

// SPDX-License-Identifier: MIT
package runtimeenv

import (
	"os"
	"runtime"
	"strings"
)

// Info describes the process runtime environment for the settings Security section.
type Info struct {
	InFlatpak         bool   `json:"inFlatpak"`
	InAppImage        bool   `json:"inAppImage"`
	InContainer       bool   `json:"inContainer"`
	ContainerRuntime  string `json:"containerRuntime,omitempty"`
	WebKitSandbox     string `json:"webkitSandbox"`
	WebKitSandboxNote string `json:"webkitSandboxNote,omitempty"`
	OnAndroid         bool   `json:"onAndroid"`
}

// Detect reports Flatpak, AppImage, container, Android, and WebKit bwrap status.
func Detect() Info {
	info := Info{
		InFlatpak:  InFlatpak(),
		InAppImage: InAppImage(),
		OnAndroid:  runtime.GOOS == "android",
	}
	info.InContainer, info.ContainerRuntime = detectContainer()
	info.WebKitSandbox, info.WebKitSandboxNote = webkitSandboxStatus(info)
	return info
}

// InFlatpak reports whether the process is running inside a Flatpak sandbox.
func InFlatpak() bool {
	if strings.TrimSpace(os.Getenv("FLATPAK_ID")) != "" {
		return true
	}
	_, err := os.Stat("/.flatpak-info")
	return err == nil
}

// InAppImage reports whether the process was launched from an AppImage.
func InAppImage() bool {
	return strings.TrimSpace(os.Getenv("APPIMAGE")) != ""
}

func detectContainer() (bool, string) {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true, "docker"
	}
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return true, "podman"
	}
	if name := strings.TrimSpace(os.Getenv("container")); name != "" {
		return true, name
	}
	data, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false, ""
	}
	text := string(data)
	for _, name := range []string{"docker", "containerd", "kubepods", "lxc", "podman", "libpod"} {
		if strings.Contains(text, name) {
			return true, name
		}
	}
	return false, ""
}

func webkitSandboxStatus(info Info) (string, string) {
	if info.OnAndroid {
		return "unavailable", "android-webview"
	}
	if runtime.GOOS != "linux" {
		return "unavailable", "not-linux"
	}
	// Container deployments are typically the server binary (no WebKitGTK).
	// Even a desktop build in Docker rarely has working nested bwrap.
	// Do not report "active" from env absence alone inside a container.
	if info.InContainer {
		return "unavailable", "container"
	}
	if strings.TrimSpace(os.Getenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS")) != "" {
		switch {
		case info.InFlatpak:
			return "disabled", "flatpak"
		case info.InAppImage:
			return "disabled", "appimage"
		default:
			return "disabled", "env"
		}
	}
	return "active", ""
}

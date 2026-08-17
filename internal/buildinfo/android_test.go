// SPDX-License-Identifier: MIT
package buildinfo

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

const shippedObtainiumVersionCode = 2

func TestAndroidVersionCode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		version string
		code    int
	}{
		{"0.2.1", 201},
		{"0.2.2", 202},
		{"0.3.0", 300},
		{"v0.2.2", 202},
		{"0.2.1-nightly.2026.08.03+2729df2", 201},
		{"0.2.1-beta.2026.08.03+abc", 201},
		{"1.0.0", 10000},
		{"0.99.99", 9999},
		{"  0.3.0  ", 300},
	}
	for _, tc := range cases {
		code, err := AndroidVersionCode(tc.version)
		if err != nil {
			t.Fatalf("AndroidVersionCode(%q) unexpected error: %v", tc.version, err)
		}
		if code != tc.code {
			t.Fatalf("AndroidVersionCode(%q) = %d, want %d", tc.version, code, tc.code)
		}
	}
}

func TestAndroidVersionCodeRejectsInvalid(t *testing.T) {
	t.Parallel()
	invalid := []string{"", "0.2", "nightly", "0.100.0", "0.2.100", "-1.0.0", "x.y.z"}
	for _, version := range invalid {
		if _, err := AndroidVersionCode(version); err == nil {
			t.Fatalf("AndroidVersionCode(%q) = nil, want error", version)
		}
	}
}

func TestAndroidVersionCodeReplacesShippedObtainiumAPKs(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"0.2.1", "0.2.2", "0.3.0"} {
		code, err := AndroidVersionCode(version)
		if err != nil {
			t.Fatalf("AndroidVersionCode(%q): %v", version, err)
		}
		if code <= shippedObtainiumVersionCode {
			t.Fatalf(
				"%s versionCode %d would not replace v0.2.1/v0.2.2 APKs shipped with versionCode %d",
				version, code, shippedObtainiumVersionCode,
			)
		}
	}
}

func TestAndroidVersionCodeMonotonic(t *testing.T) {
	t.Parallel()
	versions := []string{"0.2.1", "0.2.2", "0.3.0", "0.3.1", "1.0.0"}
	prev := 0
	for _, version := range versions {
		code, err := AndroidVersionCode(version)
		if err != nil {
			t.Fatalf("AndroidVersionCode(%q): %v", version, err)
		}
		if code <= prev {
			t.Fatalf("AndroidVersionCode(%q) = %d, want > %d", version, code, prev)
		}
		prev = code
	}
}

func TestAndroidGradleStampsVersionFromRelease(t *testing.T) {
	t.Parallel()
	gradle := readRepoFile(t, filepath.Join("build", "android", "app", "build.gradle"))
	if strings.Contains(gradle, "versionCode 2") {
		t.Fatal("hardcoded versionCode 2 was shipped for both v0.2.1 and v0.2.2")
	}
	if strings.Contains(gradle, `versionName "0.2.1"`) {
		t.Fatal("hardcoded versionName 0.2.1 was shipped for both v0.2.1 and v0.2.2")
	}
	for _, needle := range []string{
		"ANDROID_VERSION_NAME",
		"brand.yml",
		"* 10000",
		"* 100",
	} {
		if !strings.Contains(gradle, needle) {
			t.Fatalf("build.gradle must derive APK version from the release (%q missing)", needle)
		}
	}
	workflow := readRepoFile(t, filepath.Join(".github", "workflows", "desktop-build.yml"))
	if !strings.Contains(workflow, "ANDROID_VERSION_NAME: ${{ needs.version.outputs.version }}") {
		t.Fatal("desktop-build.yml must pass the resolved release version into the Android APK build")
	}
}

func TestAndroidGradleVersionCodeMatchesGo(t *testing.T) {
	t.Parallel()
	brand := readRepoFile(t, filepath.Join("build", "brand.yml"))
	match := regexp.MustCompile(`(?m)^version:\s*"([^"]+)"`).FindStringSubmatch(brand)
	if len(match) != 2 {
		t.Fatal("could not read version from build/brand.yml")
	}
	code, err := AndroidVersionCode(match[1])
	if err != nil {
		t.Fatalf("brand.yml version %q: %v", match[1], err)
	}
	if code <= shippedObtainiumVersionCode {
		t.Fatalf("brand.yml version %q yields versionCode %d, want > %d", match[1], code, shippedObtainiumVersionCode)
	}
}

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	raw, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(raw)
}

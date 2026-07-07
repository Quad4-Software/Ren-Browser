// SPDX-License-Identifier: MIT
package limits_test

import (
	"os"
	"testing"

	"renbrowser/internal/limits"
)

func TestMaxFetchBytesFilePath(t *testing.T) {
	if got := limits.MaxFetchBytes("/file/music/song.mp3"); got != 0 {
		t.Fatalf("got %d want unlimited (0)", got)
	}
}

func TestMaxFetchBytesPagePath(t *testing.T) {
	if got := limits.MaxFetchBytes("/page/index.mu"); got != limits.DefaultMaxPageBytes {
		t.Fatalf("got %d want %d", got, limits.DefaultMaxPageBytes)
	}
}

func TestEnvOverrideMaxPageBytes(t *testing.T) {
	t.Setenv("REN_BROWSER_MAX_PAGE_BYTES", "4096")
	if got := limits.MaxPageBytes(); got != 4096 {
		t.Fatalf("got %d want 4096", got)
	}
}

func TestEnvOverrideInvalidFallsBack(t *testing.T) {
	t.Setenv("REN_BROWSER_MAX_PAGE_BYTES", "nope")
	if got := limits.MaxPageBytes(); got != limits.DefaultMaxPageBytes {
		t.Fatalf("got %d want default", got)
	}
}

func TestTruncateString(t *testing.T) {
	if got := limits.TruncateString("abcdef", 3); got != "abc" {
		t.Fatalf("got %q", got)
	}
	if got := limits.TruncateString("abc", 10); got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestEnvOverrideClearsOnEmpty(t *testing.T) {
	_ = os.Unsetenv("REN_BROWSER_MAX_FILE_BYTES")
	if got := limits.MaxFileBytes(); got != 0 {
		t.Fatalf("got %d want unlimited default", got)
	}
}

func TestEnvOverrideFileBytesZeroMeansUnlimited(t *testing.T) {
	t.Setenv("REN_BROWSER_MAX_FILE_BYTES", "0")
	if got := limits.MaxFileBytes(); got != 0 {
		t.Fatalf("got %d want 0", got)
	}
}

func TestEnvOverrideFileBytesSetsLimit(t *testing.T) {
	t.Setenv("REN_BROWSER_MAX_FILE_BYTES", "4096")
	if got := limits.MaxFileBytes(); got != 4096 {
		t.Fatalf("got %d want 4096", got)
	}
}

// SPDX-License-Identifier: MIT
package sandbox

import (
	"runtime"
	"strings"
	"sync"

	"renbrowser/internal/brand"
)

// Status describes process-level filesystem sandboxing for the settings UI.
type Status struct {
	Type          string `json:"type"`
	Enabled       bool   `json:"enabled"`
	Requested     bool   `json:"requested"`
	Supported     bool   `json:"supported"`
	Auto          bool   `json:"auto"`
	DisabledByEnv bool   `json:"disabledByEnv"`
	Reason        string `json:"reason,omitempty"`
}

// Options controls whether Landlock is attempted and which paths are whitelisted.
type Options struct {
	NoLandlock      bool
	ForceLandlock   bool
	DataDir         string
	ReticulumDir    string
	ReticulumConfig string
	PluginsDir      string
	DownloadDir     string
	AssetsDir       string
	AssetsZip       string
	ExtraReadPaths  []string
}

var (
	statusMu sync.RWMutex
	status   = Status{Type: sandboxType()}
)

func sandboxType() string {
	if runtime.GOOS == "linux" {
		return "landlock"
	}
	return "none"
}

// CurrentStatus returns the last recorded sandbox status for this process.
func CurrentStatus() Status {
	statusMu.RLock()
	defer statusMu.RUnlock()
	return status
}

func setStatus(s Status) {
	statusMu.Lock()
	status = s
	statusMu.Unlock()
}

// Requested reports whether sandboxing should be attempted for the given options.
func Requested(opts Options) bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if opts.NoLandlock {
		return false
	}
	if override := envLandlockOverride(); override != nil {
		return *override
	}
	if opts.ForceLandlock {
		return true
	}
	return KernelSupported()
}

// AutoEnabled is true when sandboxing is on without an explicit env or flag override.
func AutoEnabled(opts Options) bool {
	return Requested(opts) && !opts.NoLandlock && !opts.ForceLandlock && envLandlockOverride() == nil
}

// DisabledByEnv is true when REN_BROWSER_LANDLOCK disables sandboxing.
func DisabledByEnv() bool {
	override := envLandlockOverride()
	return override != nil && !*override
}

// Apply attempts platform sandboxing. Failures are recorded in CurrentStatus and never fatal.
func Apply(opts Options) {
	s := Status{
		Type:          sandboxType(),
		Supported:     KernelSupported(),
		Requested:     Requested(opts),
		Auto:          AutoEnabled(opts),
		DisabledByEnv: DisabledByEnv(),
	}

	switch {
	case runtime.GOOS != "linux":
		s.Reason = "not supported on " + runtime.GOOS
	case opts.NoLandlock:
		s.Reason = "disabled by --no-landlock"
	case DisabledByEnv():
		s.Reason = "disabled by " + brand.EnvPrefix + "_LANDLOCK"
	case !s.Requested:
		s.Reason = "auto-disabled (kernel does not support Landlock)"
	default:
		if err := applyPlatform(opts); err != nil {
			s.Reason = "failed to apply: " + strings.TrimSpace(err.Error())
		} else {
			s.Enabled = true
		}
	}

	setStatus(s)
}

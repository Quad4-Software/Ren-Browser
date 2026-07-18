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
	Type              string `json:"type"`
	Enabled           bool   `json:"enabled"`
	Requested         bool   `json:"requested"`
	Supported         bool   `json:"supported"`
	Auto              bool   `json:"auto"`
	DisabledByEnv     bool   `json:"disabledByEnv"`
	Reason            string `json:"reason,omitempty"`
	ABI               int    `json:"abi,omitempty"`
	SeccompEnabled    bool   `json:"seccompEnabled"`
	SeccompSupported  bool   `json:"seccompSupported"`
	SeccompReason     string `json:"seccompReason,omitempty"`
	InFlatpak         bool   `json:"inFlatpak"`
	InAppImage        bool   `json:"inAppImage"`
	InContainer       bool   `json:"inContainer"`
	ContainerRuntime  string `json:"containerRuntime,omitempty"`
	WebKitSandbox     string `json:"webkitSandbox"`
	WebKitSandboxNote string `json:"webkitSandboxNote,omitempty"`
	OnAndroid         bool   `json:"onAndroid"`
}

// Options controls whether Landlock/seccomp are attempted and which paths are whitelisted.
type Options struct {
	NoLandlock      bool
	ForceLandlock   bool
	NoSeccomp       bool
	ForceSeccomp    bool
	ServerMode      bool
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

type applyResult struct {
	ABI            int
	LandlockOK     bool
	LandlockErr    error
	SeccompOK      bool
	SeccompErr     error
	SeccompSkipped bool
}

func formatApplyReason(res applyResult) string {
	var parts []string
	if res.LandlockErr != nil {
		parts = append(parts, "landlock: "+strings.TrimSpace(res.LandlockErr.Error()))
	}
	if res.SeccompErr != nil && !res.SeccompSkipped {
		parts = append(parts, "seccomp: "+strings.TrimSpace(res.SeccompErr.Error()))
	}
	if len(parts) == 0 {
		return ""
	}
	return "failed to apply: " + strings.Join(parts, "; ")
}

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
	return landlockRequested(opts) || seccompRequested(opts)
}

func landlockRequested(opts Options) bool {
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
	if !opts.ServerMode {
		return false
	}
	return KernelSupported()
}

func seccompRequested(opts Options) bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if opts.NoSeccomp {
		return false
	}
	if override := envSeccompOverride(); override != nil {
		return *override
	}
	if opts.ForceSeccomp {
		return true
	}
	// Follow the same auto policy as Landlock so desktop WebKit stays untouched
	// unless the operator opts in.
	if opts.ForceLandlock {
		return true
	}
	if !opts.ServerMode {
		return false
	}
	return SeccompSupported()
}

// AutoEnabled is true when sandboxing is on without an explicit env or flag override.
func AutoEnabled(opts Options) bool {
	return Requested(opts) &&
		!opts.NoLandlock && !opts.ForceLandlock && envLandlockOverride() == nil &&
		!opts.NoSeccomp && !opts.ForceSeccomp && envSeccompOverride() == nil
}

// DisabledByEnv is true when REN_BROWSER_LANDLOCK disables Landlock.
func DisabledByEnv() bool {
	override := envLandlockOverride()
	return override != nil && !*override
}

// Apply attempts platform sandboxing. Failures are recorded in CurrentStatus and never fatal.
func Apply(opts Options) {
	s := Status{
		Type:             sandboxType(),
		Supported:        KernelSupported(),
		Requested:        Requested(opts),
		Auto:             AutoEnabled(opts),
		DisabledByEnv:    DisabledByEnv(),
		ABI:              ABIVersion(),
		SeccompSupported: SeccompSupported(),
	}

	wantLandlock := landlockRequested(opts)
	wantSeccomp := seccompRequested(opts)

	switch {
	case runtime.GOOS != "linux":
		s.Reason = "not supported on " + runtime.GOOS
		s.SeccompReason = s.Reason
	case !wantLandlock && !wantSeccomp:
		seccompOff := false
		if o := envSeccompOverride(); o != nil && !*o {
			seccompOff = true
		}
		switch {
		case opts.NoLandlock && opts.NoSeccomp:
			s.Reason = "disabled by --no-landlock and --no-seccomp"
		case opts.NoLandlock:
			s.Reason = "disabled by --no-landlock"
		case opts.NoSeccomp:
			s.Reason = "disabled by --no-seccomp"
			s.SeccompReason = s.Reason
		case DisabledByEnv():
			s.Reason = "disabled by " + brand.EnvPrefix + "_LANDLOCK"
		case seccompOff:
			s.Reason = "disabled by " + brand.EnvPrefix + "_SECCOMP"
			s.SeccompReason = s.Reason
		case !opts.ServerMode && !opts.ForceLandlock && !opts.ForceSeccomp &&
			envLandlockOverride() == nil && envSeccompOverride() == nil:
			s.Reason = "auto-disabled on desktop (WebKitGTK nested sandbox)"
		default:
			s.Reason = "auto-disabled (kernel does not support Landlock or seccomp)"
		}
	default:
		res := applyPlatform(opts)
		s.ABI = res.ABI
		s.Enabled = res.LandlockOK
		s.SeccompEnabled = res.SeccompOK
		switch {
		case res.LandlockOK && res.SeccompOK:
			s.Type = "landlock+seccomp"
		case res.SeccompOK:
			s.Type = "seccomp"
		case res.LandlockOK:
			s.Type = "landlock"
		}
		if !res.LandlockOK && !res.SeccompOK {
			s.Reason = formatApplyReason(res)
			if s.Reason == "" {
				s.Reason = "failed to apply sandbox"
			}
		} else if reason := formatApplyReason(res); reason != "" {
			s.Reason = reason
		}
		if res.SeccompSkipped {
			s.SeccompReason = "skipped"
		} else if res.SeccompErr != nil {
			s.SeccompReason = strings.TrimSpace(res.SeccompErr.Error())
		}
	}

	setStatus(s)
}

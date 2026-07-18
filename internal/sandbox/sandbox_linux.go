// SPDX-License-Identifier: MIT

//go:build linux

package sandbox

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func applyPlatform(opts Options) applyResult {
	res := applyResult{ABI: ABIVersion()}

	if !opts.ServerMode {
		if os.Getenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS") == "" {
			_ = os.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "1")
		}
	}

	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		res.LandlockErr = fmt.Errorf("PR_SET_NO_NEW_PRIVS: %w", err)
		res.SeccompErr = res.LandlockErr
		return res
	}

	if landlockRequested(opts) {
		if err := applyLandlock(opts); err != nil {
			res.LandlockErr = err
		} else {
			res.LandlockOK = true
		}
	}

	if !seccompRequested(opts) {
		res.SeccompSkipped = true
		return res
	}

	if err := applySeccomp(); err != nil {
		// Soft-fail: seccomp is defense-in-depth and must not abort startup
		// when Landlock already applied, or when the runtime rejects filters.
		res.SeccompErr = err
		return res
	}
	res.SeccompOK = true
	return res
}

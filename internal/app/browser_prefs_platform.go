//go:build !windows

// SPDX-License-Identifier: MIT
package app

func platformDefaultNativeTitlebar() bool {
	return false
}

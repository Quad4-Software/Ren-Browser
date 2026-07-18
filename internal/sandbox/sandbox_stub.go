// SPDX-License-Identifier: MIT

//go:build !linux

package sandbox

// KernelSupported is always false off Linux.
func KernelSupported() bool {
	return false
}

// ABIVersion is always 0 off Linux.
func ABIVersion() int {
	return 0
}

// SeccompSupported is always false off Linux.
func SeccompSupported() bool {
	return false
}

func applyPlatform(opts Options) applyResult {
	return applyResult{}
}

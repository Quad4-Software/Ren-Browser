// SPDX-License-Identifier: MIT

//go:build !linux

package sandbox

// KernelSupported is always false off Linux.
func KernelSupported() bool {
	return false
}

func applyPlatform(opts Options) error {
	return nil
}

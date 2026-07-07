//go:build !android

// SPDX-License-Identifier: MIT
package rns

func usesBackboneTCPFallback() bool { return false }

// Copyright 2026 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !windows && !(linux && (amd64 || arm64 || loong64 || ppc64le || s390x || riscv64 || 386 || arm))

package libc // import "modernc.org/libc"

import (
	"os"
)

// write writes b to standard output, for printf and puts.
func write(b []byte) (int, error) {
	if _, err := os.Stdout.Write(b); err != nil {
		return -1, err
	}

	return len(b), nil
}

// setbuf is setbuf(3). The streams of this port are not buffered.
func setbuf(t *TLS, stream, buf uintptr) {}

// flushStreams writes out the output buffered by all streams, for exit. The
// streams of this port are not buffered.
func flushStreams() {}

//go:build ios

// SPDX-License-Identifier: MIT

package main

import (
	"log"
)

// handleFatalError logs a fatal startup/runtime error and parks the goroutine
// instead of calling os.Exit, which would tear down the process with no UI.
func handleFatalError(err error) {
	log.Printf("FATAL ERROR: %v", err)
	select {}
}

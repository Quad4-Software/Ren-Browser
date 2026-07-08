//go:build !ios

// SPDX-License-Identifier: MIT

package main

import (
	"log"
)

func handleFatalError(err error) {
	log.Fatal(err)
}

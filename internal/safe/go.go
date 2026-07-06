// SPDX-License-Identifier: MIT
package safe

import (
	"log"
	"runtime/debug"
)

func Go(name string, fn func()) {
	go func() {
		defer recoverAndLog(name)
		fn()
	}()
}

func recoverAndLog(name string) {
	if r := recover(); r != nil {
		log.Printf("%s panic: %v\n%s", name, r, debug.Stack())
	}
}

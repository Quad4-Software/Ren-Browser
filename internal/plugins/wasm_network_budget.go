// SPDX-License-Identifier: MIT
package plugins

import (
	"fmt"
	"sync"
)

const wasmMaxFetchesPerCall = 64

type wasmFetchBudget struct {
	max     int
	used    int
	aborted bool
}

var wasmFetchBudgets sync.Map

func beginWasmFetchBudget(pluginID string, max int) {
	if max <= 0 {
		max = wasmMaxFetchesPerCall
	}
	wasmFetchBudgets.Store(pluginID, &wasmFetchBudget{max: max})
}

func endWasmFetchBudget(pluginID string) {
	wasmFetchBudgets.Delete(pluginID)
}

func consumeWasmFetchBudget(pluginID string) error {
	raw, ok := wasmFetchBudgets.Load(pluginID)
	if !ok {
		return nil
	}
	budget := raw.(*wasmFetchBudget)
	budget.used++
	if budget.used > budget.max {
		budget.aborted = true
		return fmt.Errorf("extension exceeded network request limit (%d)", budget.max)
	}
	return nil
}

func wasmExportNeedsNetwork(exportName string) bool {
	switch exportName {
	case "translate_micron":
		return true
	default:
		return false
	}
}

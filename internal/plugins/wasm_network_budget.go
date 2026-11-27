// SPDX-License-Identifier: MIT
package plugins

import (
	"context"
	"fmt"
	"sync"
)

const wasmMaxFetchesPerCall = 64

type wasmFetchBudget struct {
	mu   sync.Mutex
	max  int
	used int
}

type wasmFetchBudgetContextKey struct{}

func newWasmFetchBudget(max int) *wasmFetchBudget {
	if max <= 0 {
		max = wasmMaxFetchesPerCall
	}
	return &wasmFetchBudget{max: max}
}

func withWasmFetchBudget(ctx context.Context, budget *wasmFetchBudget) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, wasmFetchBudgetContextKey{}, budget)
}

func consumeWasmFetchBudget(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	budget, ok := ctx.Value(wasmFetchBudgetContextKey{}).(*wasmFetchBudget)
	if !ok || budget == nil {
		return nil
	}
	budget.mu.Lock()
	defer budget.mu.Unlock()
	budget.used++
	if budget.used > budget.max {
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

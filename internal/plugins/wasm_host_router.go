// SPDX-License-Identifier: MIT
package plugins

import (
	"context"
	"sync"

	"github.com/tetratelabs/wazero/api"
)

type wasmHostRouter struct {
	mu      sync.Mutex
	plugins map[string]*wasmHost
}

func newWasmHostRouter() *wasmHostRouter {
	return &wasmHostRouter{
		plugins: make(map[string]*wasmHost),
	}
}

func (r *wasmHostRouter) register(pluginID string, host *wasmHost) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[pluginID] = host
}

func (r *wasmHostRouter) unregister(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.plugins, pluginID)
}

func (r *wasmHostRouter) httpFetch(ctx context.Context, mod api.Module, reqPtr, reqLen, respPtr, respCap uint32) uint32 {
	r.mu.Lock()
	host := r.plugins[mod.Name()]
	r.mu.Unlock()
	if host == nil {
		return writeWasmHTTPError(mod, respPtr, respCap, "plugin wasm host not registered")
	}
	return host.httpFetch(ctx, mod, reqPtr, reqLen, respPtr, respCap)
}

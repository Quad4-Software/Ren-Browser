// SPDX-License-Identifier: MIT
package plugins

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/tetratelabs/wazero"
)

type WasmRuntime struct {
	mu         sync.Mutex
	runtime    wazero.Runtime
	modules    map[string]*wasmPlugin
	hostRouter *wasmHostRouter
	hostOnce   sync.Once
	hostInit   error
}

func NewWasmRuntime() *WasmRuntime {
	return &WasmRuntime{
		runtime: wazero.NewRuntime(context.Background()),
		modules: make(map[string]*wasmPlugin),
	}
}

func (rt *WasmRuntime) ensureHostModule(ctx context.Context) error {
	rt.hostOnce.Do(func() {
		rt.hostRouter = newWasmHostRouter()
		_, rt.hostInit = rt.runtime.NewHostModuleBuilder(wasmHostModule).
			NewFunctionBuilder().
			WithFunc(rt.hostRouter.httpFetch).
			Export("http_fetch").
			Instantiate(ctx)
	})
	return rt.hostInit
}

func (rt *WasmRuntime) LoadPlugin(pluginID, wasmPath string, manifest Manifest, granted []string) (*wasmPlugin, error) {
	return rt.LoadPluginWithFetch(pluginID, wasmPath, manifest, granted, nil)
}

func (rt *WasmRuntime) LoadPluginWithFetch(pluginID, wasmPath string, manifest Manifest, granted []string, fetch func(WasmHTTPRequest) (WasmHTTPResponse, error)) (*wasmPlugin, error) {
	data, err := os.ReadFile(wasmPath) // #nosec G304 -- plugin wasm from managed dir
	if err != nil {
		return nil, err
	}
	return rt.loadPlugin(pluginID, data, manifest, granted, fetch)
}

func (rt *WasmRuntime) loadPlugin(pluginID string, data []byte, manifest Manifest, granted []string, fetch func(WasmHTTPRequest) (WasmHTTPResponse, error)) (*wasmPlugin, error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if existing, ok := rt.modules[pluginID]; ok {
		return existing, nil
	}

	networkFetch := HasGrantedPermission(granted, PermNetworkFetch)
	host := newWasmHost(pluginID, manifest, networkFetch, fetch)
	ctx, cancel := context.WithTimeout(context.Background(), wasmLoadTimeout)
	defer cancel()

	if err := rt.ensureHostModule(ctx); err != nil {
		return nil, fmt.Errorf("instantiate wasm host: %w", err)
	}
	rt.hostRouter.register(pluginID, host)

	mod, err := rt.runtime.InstantiateWithConfig(ctx, data, wazero.NewModuleConfig().WithName(pluginID))
	if err != nil {
		rt.hostRouter.unregister(pluginID)
		return nil, fmt.Errorf("instantiate wasm plugin: %w", err)
	}
	if init := mod.ExportedFunction("_initialize"); init != nil {
		if _, err := init.Call(ctx); err != nil {
			_ = mod.Close(context.Background())
			rt.hostRouter.unregister(pluginID)
			return nil, fmt.Errorf("wasm initialize: %w", err)
		}
	}

	wp := &wasmPlugin{
		pluginID: pluginID,
		manifest: manifest,
		mod:      mod,
		host:     host,
	}
	rt.modules[pluginID] = wp
	return wp, nil
}

func (rt *WasmRuntime) Get(pluginID string) (*wasmPlugin, bool) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	wp, ok := rt.modules[pluginID]
	return wp, ok
}

func (rt *WasmRuntime) LoadRenderer(pluginID, wasmPath string, manifest Manifest, granted []string) (Renderer, error) {
	wp, err := rt.LoadPlugin(pluginID, wasmPath, manifest, granted)
	if err != nil {
		return nil, err
	}
	return wasmRendererAdapter{plugin: wp}, nil
}

func (rt *WasmRuntime) Close() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	for id, wp := range rt.modules {
		_ = wp.mod.Close(context.Background())
		delete(rt.modules, id)
	}
	if rt.hostRouter != nil {
		rt.hostRouter = newWasmHostRouter()
	}
	rt.hostOnce = sync.Once{}
	rt.hostInit = nil
	if rt.runtime != nil {
		return rt.runtime.Close(context.Background())
	}
	return nil
}

func (rt *WasmRuntime) Unload(pluginID string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if wp, ok := rt.modules[pluginID]; ok {
		_ = wp.mod.Close(context.Background())
		delete(rt.modules, pluginID)
	}
	if rt.hostRouter != nil {
		rt.hostRouter.unregister(pluginID)
	}
}

func (rt *WasmRuntime) BeforeFetchHook(pluginID string) BeforeFetchHook {
	wp, ok := rt.Get(pluginID)
	if !ok {
		return func(ctx FetchContext) (FetchHookResult, error) {
			return FetchHookResult{}, nil
		}
	}
	return wp.BeforeFetchHook()
}

func (rt *WasmRuntime) AfterFetchHook(pluginID string) AfterFetchHook {
	wp, ok := rt.Get(pluginID)
	if !ok {
		return func(ctx FetchContext, body []byte) ([]byte, error) {
			return body, nil
		}
	}
	return wp.AfterFetchHook()
}

func (rt *WasmRuntime) loadPluginForTest(pluginID string, data []byte, manifest Manifest, granted []string, fetch func(WasmHTTPRequest) (WasmHTTPResponse, error)) (*wasmPlugin, error) {
	rt.mu.Lock()
	if wp, ok := rt.modules[pluginID]; ok {
		rt.mu.Unlock()
		return wp, nil
	}
	rt.mu.Unlock()
	return rt.loadPlugin(pluginID, data, manifest, granted, fetch)
}

// LoadPluginForTest loads a wasm plugin with an optional HTTP fetch stub.
func (rt *WasmRuntime) LoadPluginForTest(pluginID string, data []byte, manifest Manifest, fetch func(WasmHTTPRequest) (WasmHTTPResponse, error)) (*wasmPlugin, error) {
	return rt.loadPluginForTest(pluginID, data, manifest, DefaultGrantedPermissions(manifest), fetch)
}

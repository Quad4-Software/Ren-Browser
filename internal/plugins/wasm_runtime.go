// SPDX-License-Identifier: MIT
package plugins

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const wasmTimeout = 500 * time.Millisecond

type WasmRuntime struct {
	mu      sync.Mutex
	runtime wazero.Runtime
	modules map[string]api.Module
}

func NewWasmRuntime() *WasmRuntime {
	return &WasmRuntime{
		runtime: wazero.NewRuntime(context.Background()),
		modules: make(map[string]api.Module),
	}
}

func (rt *WasmRuntime) LoadRenderer(pluginID, wasmPath string, manifest Manifest) (Renderer, error) {
	data, err := os.ReadFile(wasmPath) // #nosec G304 -- plugin wasm from managed dir
	if err != nil {
		return nil, err
	}
	mod, err := rt.loadModule(pluginID, data)
	if err != nil {
		return nil, err
	}
	return &wasmRenderer{pluginID: pluginID, manifest: manifest, mod: mod}, nil
}

func (rt *WasmRuntime) loadModule(pluginID string, data []byte) (api.Module, error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if mod, ok := rt.modules[pluginID]; ok {
		return mod, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mod, err := rt.runtime.InstantiateWithConfig(ctx, data, wazero.NewModuleConfig())
	if err != nil {
		return nil, err
	}
	rt.modules[pluginID] = mod
	return mod, nil
}

func (rt *WasmRuntime) Close() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	for id, mod := range rt.modules {
		_ = mod.Close(context.Background())
		delete(rt.modules, id)
	}
	if rt.runtime != nil {
		return rt.runtime.Close(context.Background())
	}
	return nil
}

func (rt *WasmRuntime) Unload(pluginID string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if mod, ok := rt.modules[pluginID]; ok {
		_ = mod.Close(context.Background())
		delete(rt.modules, pluginID)
	}
}

func (rt *WasmRuntime) BeforeFetchHook(pluginID string) BeforeFetchHook {
	return func(ctx FetchContext) (FetchHookResult, error) {
		return FetchHookResult{}, nil
	}
}

func (rt *WasmRuntime) AfterFetchHook(pluginID string) AfterFetchHook {
	return func(ctx FetchContext, body []byte) ([]byte, error) {
		return body, nil
	}
}

type wasmRenderer struct {
	pluginID string
	manifest Manifest
	mod      api.Module
}

func (w *wasmRenderer) ID() string {
	for _, r := range w.manifest.Contributes.Renderers {
		return w.pluginID + "." + r.ID
	}
	return w.pluginID + ".wasm"
}

func (w *wasmRenderer) Priority() int {
	for _, r := range w.manifest.Contributes.Renderers {
		if r.Priority > 0 {
			return r.Priority
		}
	}
	return 50
}

func (w *wasmRenderer) PluginID() string { return w.pluginID }

func (w *wasmRenderer) Match(path string, body []byte, detected string) bool {
	for _, r := range w.manifest.Contributes.Renderers {
		for _, ext := range r.Extensions {
			if ext != "" && hasSuffixFold(path, ext) {
				return true
			}
		}
		for _, mime := range r.MIME {
			if mime != "" && detected == mime {
				return true
			}
		}
	}
	return false
}

func (w *wasmRenderer) Render(path string, body []byte, nodeHash string) (Rendered, error) {
	fn := w.mod.ExportedFunction("render")
	if fn == nil {
		return Rendered{}, fmt.Errorf("wasm module missing render export")
	}
	ctx, cancel := context.WithTimeout(context.Background(), wasmTimeout)
	defer cancel()
	_, err := fn.Call(ctx,
		api.EncodeU32(uint32(len(path))),     // #nosec G115 -- wasm ABI length hint
		api.EncodeU32(uint32(len(body))),     // #nosec G115 -- wasm ABI length hint
		api.EncodeU32(uint32(len(nodeHash))), // #nosec G115 -- wasm ABI length hint
	)
	if err != nil {
		return Rendered{}, fmt.Errorf("wasm render failed: %w", err)
	}
	htmlOut := fmt.Sprintf("<pre class=\"plaintext\">wasm render stub (%d bytes)</pre>", len(body))
	htmlOut = SanitizePluginHTML(htmlOut, w.manifest)
	return Rendered{Kind: "html", HTML: htmlOut, Raw: string(body)}, nil
}

func hasSuffixFold(path, suffix string) bool {
	if len(path) < len(suffix) {
		return false
	}
	return equalFold(path[len(path)-len(suffix):], suffix)
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

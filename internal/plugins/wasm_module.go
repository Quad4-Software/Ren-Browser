// SPDX-License-Identifier: MIT
package plugins

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

type wasmPlugin struct {
	pluginID string
	manifest Manifest
	mod      api.Module
	host     *wasmHost
}

func (wp *wasmPlugin) PluginID() string { return wp.pluginID }

func (wp *wasmPlugin) hasExport(name string) bool {
	return wp.mod.ExportedFunction(name) != nil
}

func (wp *wasmPlugin) CallExport(export string, input []byte) ([]byte, error) {
	if len(input) > wasmMaxIOBytes {
		return nil, fmt.Errorf("wasm input too large")
	}
	fn := wp.mod.ExportedFunction(export)
	if fn == nil {
		return nil, fmt.Errorf("wasm module missing export %q", export)
	}
	mem := wp.mod.Memory()
	if mem == nil {
		return nil, fmt.Errorf("wasm module missing memory export")
	}
	scratch := make([]byte, len(input))
	copy(scratch, input)
	if !mem.Write(wasmScratchAddr, scratch) {
		return nil, fmt.Errorf("wasm memory write failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), wasmCallTimeout)
	defer cancel()
	results, err := fn.Call(ctx, api.EncodeU32(wasmScratchAddr), api.EncodeU32(uint32(len(input)))) // #nosec G115 -- wasm ABI
	if err != nil {
		return nil, fmt.Errorf("wasm %s failed: %w", export, err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("wasm %s returned no length", export)
	}
	outLen := results[0]
	if outLen == 0 {
		return []byte{}, nil
	}
	if outLen > wasmMaxIOBytes {
		return nil, fmt.Errorf("wasm output too large")
	}
	out, ok := mem.Read(wasmScratchAddr, uint32(outLen)) // #nosec G115 -- wasm ABI
	if !ok {
		return nil, fmt.Errorf("wasm memory read failed")
	}
	copied := make([]byte, len(out))
	copy(copied, out)
	return copied, nil
}

func (wp *wasmPlugin) CallJSON(export string, input []byte) ([]byte, error) {
	return wp.CallExport(export, input)
}

func (wp *wasmPlugin) AfterFetchHook() AfterFetchHook {
	return func(ctx FetchContext, body []byte) ([]byte, error) {
		if !wp.hasExport("after_fetch") {
			return body, nil
		}
		req, err := jsonMarshal(WasmTranslateRequest{
			Body: string(body),
			Path: ctx.Path,
		})
		if err != nil {
			return body, err
		}
		out, err := wp.CallExport("after_fetch", req)
		if err != nil {
			return body, err
		}
		var resp WasmTranslateResponse
		if err := json.Unmarshal(out, &resp); err != nil {
			return body, err
		}
		if resp.Error != "" {
			return body, fmt.Errorf("%s", resp.Error)
		}
		if resp.Body == "" {
			return body, nil
		}
		return []byte(resp.Body), nil
	}
}

func (wp *wasmPlugin) BeforeFetchHook() BeforeFetchHook {
	return func(ctx FetchContext) (FetchHookResult, error) {
		if !wp.hasExport("before_fetch") {
			return FetchHookResult{}, nil
		}
		req, err := jsonMarshal(struct {
			Path     string `json:"path"`
			NodeHash string `json:"nodeHash"`
		}{
			Path:     ctx.Path,
			NodeHash: ctx.NodeHash,
		})
		if err != nil {
			return FetchHookResult{}, err
		}
		out, err := wp.CallExport("before_fetch", req)
		if err != nil {
			return FetchHookResult{}, err
		}
		var resp struct {
			Cancel bool   `json:"cancel"`
			Path   string `json:"path"`
			Error  string `json:"error"`
		}
		if err := json.Unmarshal(out, &resp); err != nil {
			return FetchHookResult{}, err
		}
		if resp.Error != "" {
			return FetchHookResult{}, fmt.Errorf("%s", resp.Error)
		}
		return FetchHookResult{Cancel: resp.Cancel, Path: resp.Path}, nil
	}
}

type wasmRendererAdapter struct {
	plugin *wasmPlugin
}

func (w wasmRendererAdapter) ID() string {
	for _, r := range w.plugin.manifest.Contributes.Renderers {
		return w.plugin.pluginID + "." + r.ID
	}
	return w.plugin.pluginID + ".wasm"
}

func (w wasmRendererAdapter) Priority() int {
	for _, r := range w.plugin.manifest.Contributes.Renderers {
		if r.Priority > 0 {
			return r.Priority
		}
	}
	return 50
}

func (w wasmRendererAdapter) PluginID() string { return w.plugin.pluginID }

func (w wasmRendererAdapter) Match(path string, body []byte, detected string) bool {
	for _, r := range w.plugin.manifest.Contributes.Renderers {
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

func (w wasmRendererAdapter) Render(path string, body []byte, nodeHash string) (Rendered, error) {
	req, err := jsonMarshal(struct {
		Path     string `json:"path"`
		Body     string `json:"body"`
		NodeHash string `json:"nodeHash"`
	}{
		Path:     path,
		Body:     string(body),
		NodeHash: nodeHash,
	})
	if err != nil {
		return Rendered{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), wasmRenderBudget)
	defer cancel()
	_ = ctx

	out, err := w.plugin.CallExport("render", req)
	if err != nil {
		return Rendered{}, err
	}
	var resp struct {
		HTML   string `json:"html"`
		Kind   string `json:"kind"`
		Raw    string `json:"raw"`
		PageFG string `json:"pageFg"`
		PageBG string `json:"pageBg"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return Rendered{}, err
	}
	if resp.Error != "" {
		return Rendered{}, fmt.Errorf("%s", resp.Error)
	}
	htmlOut := SanitizePluginHTML(resp.HTML, w.plugin.manifest)
	kind := resp.Kind
	if kind == "" {
		kind = "html"
	}
	raw := resp.Raw
	if raw == "" {
		raw = string(body)
	}
	return Rendered{
		Kind:   kind,
		HTML:   htmlOut,
		Raw:    raw,
		PageFG: resp.PageFG,
		PageBG: resp.PageBG,
	}, nil
}

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
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

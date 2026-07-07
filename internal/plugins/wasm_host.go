// SPDX-License-Identifier: MIT
package plugins

import (
	"context"
	"encoding/json"

	"github.com/tetratelabs/wazero/api"
)

type wasmHost struct {
	pluginID     string
	manifest     Manifest
	networkFetch bool
	fetch        func(WasmHTTPRequest) (WasmHTTPResponse, error)
}

func newWasmHost(pluginID string, manifest Manifest, networkFetch bool, fetch func(WasmHTTPRequest) (WasmHTTPResponse, error)) *wasmHost {
	if fetch == nil {
		fetch = DoPluginHTTP
	}
	return &wasmHost{
		pluginID:     pluginID,
		manifest:     manifest,
		networkFetch: networkFetch,
		fetch:        fetch,
	}
}

func (h *wasmHost) httpFetch(ctx context.Context, mod api.Module, reqPtr, reqLen, respPtr, respCap uint32) uint32 {
	if !h.networkFetch {
		return h.writeHTTPError(mod, respPtr, respCap, "plugin lacks network.fetch permission")
	}
	reqBytes, ok := mod.Memory().Read(reqPtr, reqLen)
	if !ok {
		return h.writeHTTPError(mod, respPtr, respCap, "invalid request memory")
	}
	var req WasmHTTPRequest
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		return h.writeHTTPError(mod, respPtr, respCap, "invalid request JSON")
	}
	resp, err := h.fetch(req)
	if err != nil {
		return h.writeHTTPError(mod, respPtr, respCap, err.Error())
	}
	out, err := json.Marshal(resp)
	if err != nil {
		return h.writeHTTPError(mod, respPtr, respCap, "marshal response failed")
	}
	if len(out) > int(respCap) {
		return h.writeHTTPError(mod, respPtr, respCap, "response buffer too small")
	}
	if !mod.Memory().Write(respPtr, out) {
		return h.writeHTTPError(mod, respPtr, respCap, "write response failed")
	}
	return uint32(len(out)) // #nosec G115 -- wasm ABI length
}

func (h *wasmHost) writeHTTPError(mod api.Module, respPtr, respCap uint32, message string) uint32 {
	out, err := json.Marshal(WasmHTTPResponse{Error: message})
	if err != nil || len(out) > int(respCap) {
		return 0
	}
	if !mod.Memory().Write(respPtr, out) {
		return 0
	}
	return uint32(len(out)) // #nosec G115 -- wasm ABI length
}

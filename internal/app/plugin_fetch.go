// SPDX-License-Identifier: MIT
package app

import (
	"fmt"

	"renbrowser/internal/plugins"
)

type PluginHTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type PluginHTTPResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func (h *PluginHost) PluginFetch(pluginID string, req PluginHTTPRequest) (PluginHTTPResponse, error) {
	if h.manager == nil {
		return PluginHTTPResponse{}, fmt.Errorf("plugin host unavailable")
	}
	p, ok := h.manager.Get(pluginID)
	if !ok || !p.Enabled {
		return PluginHTTPResponse{}, fmt.Errorf("plugin %q not enabled", pluginID)
	}
	if err := plugins.RequireGrantedPermission(p.GrantedPermissions, p.Manifest, plugins.PermNetworkFetch); err != nil {
		return PluginHTTPResponse{}, err
	}

	resp, err := h.manager.PluginHTTPFetch(pluginID, plugins.WasmHTTPRequest{
		Method:  req.Method,
		URL:     req.URL,
		Headers: req.Headers,
		Body:    req.Body,
	})
	if err != nil {
		h.manager.LogPluginError(pluginID, "network.fetch", err.Error(), "")
		return PluginHTTPResponse{}, err
	}
	return PluginHTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
	}, nil
}

func (h *PluginHost) PluginWasmCall(pluginID, exportName, input string) (string, error) {
	if h.manager == nil {
		return "", fmt.Errorf("plugin host unavailable")
	}
	return h.manager.WasmCall(pluginID, exportName, input)
}

func (h *PluginHost) ReportPluginError(pluginID, phase, message, detail string) error {
	if h.manager == nil {
		return fmt.Errorf("plugin host unavailable")
	}
	if detail != "" && detail != message {
		h.manager.LogPluginError(pluginID, phase, message, detail)
	}
	return h.manager.FailPlugin(pluginID, phase, fmt.Errorf("%s", message))
}

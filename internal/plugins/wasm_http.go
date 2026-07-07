// SPDX-License-Identifier: MIT
package plugins

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const pluginHTTPTimeout = 30 * time.Second
const pluginFetchMaxBody = 1 << 20
const pluginFetchMaxURLLen = 2048

func DoPluginHTTP(req WasmHTTPRequest) (WasmHTTPResponse, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	if method != http.MethodGet && method != http.MethodPost {
		return WasmHTTPResponse{}, fmt.Errorf("unsupported HTTP method %q", method)
	}

	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" || len(rawURL) > pluginFetchMaxURLLen {
		return WasmHTTPResponse{}, fmt.Errorf("invalid request URL")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return WasmHTTPResponse{}, fmt.Errorf("invalid request URL: %w", err)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return WasmHTTPResponse{}, fmt.Errorf("only http and https URLs are allowed")
	}

	var body io.Reader
	if req.Body != "" {
		if len(req.Body) > pluginFetchMaxBody {
			return WasmHTTPResponse{}, fmt.Errorf("request body too large")
		}
		body = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return WasmHTTPResponse{}, err
	}
	for key, value := range req.Headers {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		httpReq.Header.Set(key, value)
	}
	if method == http.MethodPost && httpReq.Header.Get("Content-Type") == "" && req.Body != "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: pluginHTTPTimeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return WasmHTTPResponse{}, err
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, pluginFetchMaxBody+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return WasmHTTPResponse{}, err
	}
	if len(data) > pluginFetchMaxBody {
		return WasmHTTPResponse{}, fmt.Errorf("response body too large")
	}

	return WasmHTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       string(data),
	}, nil
}

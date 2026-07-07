// SPDX-License-Identifier: MIT
package plugins

import "time"

const (
	wasmHostModule   = "renhost"
	wasmScratchAddr  = 4096
	wasmMaxIOBytes   = 1 << 20
	wasmCallTimeout  = 30 * time.Second
	wasmLoadTimeout  = 5 * time.Second
	wasmRenderBudget = 500 * time.Millisecond
)

type WasmTranslateRequest struct {
	Body     string                `json:"body"`
	Path     string                `json:"path,omitempty"`
	Settings WasmTranslateSettings `json:"settings"`
}

type WasmTranslateSettings struct {
	Backend              string `json:"backend"`
	TargetLang           string `json:"targetLang"`
	SourceLang           string `json:"sourceLang"`
	LibretranslateURL    string `json:"libretranslateUrl"`
	LibretranslateAPIKey string `json:"libretranslateApiKey"`
}

type WasmTranslateResponse struct {
	Body  string `json:"body"`
	Error string `json:"error,omitempty"`
}

type WasmHTTPRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type WasmHTTPResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
	Error      string `json:"error,omitempty"`
}

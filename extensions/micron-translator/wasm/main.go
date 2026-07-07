//go:build wasm

// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"net/url"
	"strings"
	"unicode"
	"unsafe"
)

const (
	scratchAddr            = 4096
	httpReqAddr            = 8192
	httpRespAddr           = 16384
	httpRespCap            = 32768
	backendGoogle          = "google"
	maxTranslateSegments   = 128
	maxTranslateInputBytes = 512 * 1024
)

//go:wasmimport renhost http_fetch
func httpFetch(reqPtr, reqLen, respPtr, respCap uint32) uint32

type translateRequest struct {
	Body     string            `json:"body"`
	Settings translateSettings `json:"settings"`
}

type translateSettings struct {
	Backend              string `json:"backend"`
	TargetLang           string `json:"targetLang"`
	SourceLang           string `json:"sourceLang"`
	LibretranslateURL    string `json:"libretranslateUrl"`
	LibretranslateAPIKey string `json:"libretranslateApiKey"`
}

type translateResponse struct {
	Body  string `json:"body"`
	Error string `json:"error,omitempty"`
}

type httpRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type httpResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
	Error      string `json:"error,omitempty"`
}

type segment struct {
	markup bool
	value  string
}

//export translate_micron
func translateMicron(inPtr, inLen uint32) uint32 {
	input := memSlice(inPtr, inLen)
	if len(input) > maxTranslateInputBytes {
		return writeResponse(translateResponse{Error: "page content too large to translate"})
	}
	var req translateRequest
	if err := json.Unmarshal(input, &req); err != nil {
		return writeResponse(translateResponse{Error: "invalid request JSON"})
	}
	translated, err := translateBody(req.Body, req.Settings)
	if err != nil {
		return writeResponse(translateResponse{Error: err.Error()})
	}
	return writeResponse(translateResponse{Body: translated})
}

func writeResponse(resp translateResponse) uint32 {
	out, err := json.Marshal(resp)
	if err != nil {
		out = []byte(`{"error":"marshal response failed"}`)
	}
	copy(memSlice(scratchAddr, uint32(len(out))), out) // #nosec G115 -- wasm ABI
	return uint32(len(out))                            // #nosec G115 -- wasm ABI
}

func translateBody(source string, settings translateSettings) (string, error) {
	segments := splitSegments(source)
	translatable := 0
	for _, seg := range segments {
		if !seg.markup && isTranslatable(seg.value) {
			translatable++
		}
	}
	if translatable > maxTranslateSegments {
		return "", errTooManySegments()
	}
	var b strings.Builder
	for _, seg := range segments {
		if seg.markup || !isTranslatable(seg.value) {
			b.WriteString(seg.value)
			continue
		}
		translated, err := translateText(seg.value, settings)
		if err != nil {
			return "", err
		}
		b.WriteString(translated)
	}
	return b.String(), nil
}

func splitSegments(source string) []segment {
	out := make([]segment, 0, 8)
	i := 0
	for i < len(source) {
		if source[i] == '`' {
			j := i + 1
			for j < len(source) && source[j] != '`' {
				j++
			}
			if j < len(source) {
				out = append(out, segment{markup: true, value: source[i : j+1]})
				i = j + 1
				continue
			}
		}
		j := i
		for j < len(source) && source[j] != '`' {
			j++
		}
		if chunk := source[i:j]; chunk != "" {
			out = append(out, segment{value: chunk})
		}
		i = j
	}
	return out
}

func isTranslatable(text string) bool {
	for _, r := range strings.TrimSpace(text) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return true
		}
	}
	return false
}

func translateText(text string, settings translateSettings) (string, error) {
	if settings.Backend == backendGoogle || settings.Backend == "" {
		return translateGoogle(text, settings)
	}
	return translateLibre(text, settings)
}

func translateGoogle(text string, settings translateSettings) (string, error) {
	source := settings.SourceLang
	if source == "" {
		source = "auto"
	}
	target := settings.TargetLang
	if target == "" {
		target = "en"
	}
	q := url.QueryEscape(text)
	rawURL := "https://translate.googleapis.com/translate_a/single?client=gtx&sl=" +
		url.QueryEscape(source) + "&tl=" + url.QueryEscape(target) + "&dt=t&q=" + q
	resp, err := hostHTTP(httpRequest{Method: "GET", URL: rawURL})
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errHTTPStatus(resp.StatusCode)
	}
	return parseGoogleBody(resp.Body)
}

func translateLibre(text string, settings translateSettings) (string, error) {
	base := strings.TrimRight(settings.LibretranslateURL, "/")
	if base == "" {
		base = "https://libretranslate.com"
	}
	source := settings.SourceLang
	if source == "" {
		source = "auto"
	}
	target := settings.TargetLang
	if target == "" {
		target = "en"
	}
	payload := map[string]string{
		"q":      text,
		"source": source,
		"target": target,
		"format": "text",
	}
	if settings.LibretranslateAPIKey != "" {
		payload["api_key"] = settings.LibretranslateAPIKey
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	resp, err := hostHTTP(httpRequest{
		Method:  "POST",
		URL:     base + "/translate",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    string(body),
	})
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errHTTPStatus(resp.StatusCode)
	}
	var parsed struct {
		TranslatedText string `json:"translatedText"`
	}
	if err := json.Unmarshal([]byte(resp.Body), &parsed); err != nil {
		return "", err
	}
	if parsed.TranslatedText == "" {
		return "", errMissingTranslation()
	}
	return parsed.TranslatedText, nil
}

func parseGoogleBody(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", errMissingTranslation()
	}
	if strings.HasPrefix(body, ")]}'") {
		if idx := strings.Index(body, "["); idx >= 0 {
			body = body[idx:]
		}
	}
	if strings.HasPrefix(body, "\"") {
		var msg string
		if err := json.Unmarshal([]byte(body), &msg); err == nil {
			if msg == "" {
				return "", errMissingTranslation()
			}
			return "", errString(msg)
		}
	}
	if strings.HasPrefix(body, "<") {
		return "", errHTTPStatus(502)
	}
	var data []interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errMissingTranslation()
	}
	segments, ok := data[0].([]interface{})
	if !ok {
		return "", errMissingTranslation()
	}
	var b strings.Builder
	for _, part := range segments {
		row, ok := part.([]interface{})
		if !ok || len(row) == 0 {
			continue
		}
		if s, ok := row[0].(string); ok {
			b.WriteString(s)
		}
	}
	if b.Len() == 0 {
		return "", errMissingTranslation()
	}
	return b.String(), nil
}

func hostHTTP(req httpRequest) (httpResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return httpResponse{}, err
	}
	reqPtr := uint32(httpReqAddr)                                                  // #nosec G115 -- wasm scratch layout
	copy(memSlice(reqPtr, uint32(len(reqBytes))), reqBytes)                        // #nosec G115 -- wasm ABI
	written := httpFetch(reqPtr, uint32(len(reqBytes)), httpRespAddr, httpRespCap) // #nosec G115 -- wasm ABI
	if written == 0 {
		return httpResponse{}, errHostFetch()
	}
	var resp httpResponse
	if err := json.Unmarshal(memSlice(httpRespAddr, written), &resp); err != nil {
		return httpResponse{}, err
	}
	if resp.Error != "" {
		return httpResponse{}, errString(resp.Error)
	}
	return resp, nil
}

func memSlice(ptr, ln uint32) []byte {
	if ln == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), int(ln)) // #nosec G103 -- wasm linear memory
}

type simpleError string

func (e simpleError) Error() string { return string(e) }

func errHTTPStatus(code int) error {
	return simpleError("translation HTTP error")
}

func errMissingTranslation() error {
	return simpleError("translation missing")
}

func errHostFetch() error {
	return simpleError("host fetch failed")
}

func errTooManySegments() error {
	return simpleError("page has too many text segments to translate safely")
}

func errString(msg string) error {
	return simpleError(msg)
}

func main() {}

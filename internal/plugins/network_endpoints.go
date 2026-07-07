// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	urlInTextRe  = regexp.MustCompile(`https?://[^\s"'<>\\)\]]+`)
	schemeHostRe = regexp.MustCompile(`https?://([a-z0-9][-a-z0-9.]*(?:\.[a-z0-9][-a-z0-9.]*)+)`)
)

func CollectNetworkEndpoints(manifest Manifest, dir string, embedded map[string][]byte) []string {
	seen := make(map[string]struct{})
	var manifestEndpoints []string
	var scanned []string
	add := func(value string, requireHTTP bool, manifestDeclared bool) {
		value = normalizeEndpoint(value)
		if value == "" {
			return
		}
		if requireHTTP && !isHTTPURL(value) {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		if manifestDeclared {
			manifestEndpoints = append(manifestEndpoints, value)
			return
		}
		scanned = append(scanned, value)
	}

	if manifest.Network != nil {
		for _, endpoint := range manifest.Network.Endpoints {
			urls := extractURLsFromText(endpoint)
			if len(urls) == 0 {
				add(endpoint, false, true)
				continue
			}
			for _, value := range urls {
				add(value, false, true)
			}
		}
	}

	for path, data := range embedded {
		if !shouldScanNetworkFile(path) {
			continue
		}
		for _, endpoint := range extractURLsFromBytes(data) {
			add(endpoint, true, false)
		}
	}

	if dir != "" {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if !shouldScanNetworkFile(path) {
				return nil
			}
			data, readErr := os.ReadFile(path) // #nosec G304 -- plugin preview dir from validated extract
			if readErr != nil {
				return nil
			}
			for _, endpoint := range extractURLsFromBytes(data) {
				add(endpoint, true, false)
			}
			return nil
		})
	}

	for _, endpoint := range append([]string{}, manifestEndpoints...) {
		addHostRoots(endpoint, seen, func(value string) {
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
			scanned = append(scanned, value)
		})
	}
	for _, endpoint := range append([]string{}, scanned...) {
		addHostRoots(endpoint, seen, func(value string) {
			if _, ok := seen[value]; ok {
				return
			}
			seen[value] = struct{}{}
			scanned = append(scanned, value)
		})
	}

	sort.Strings(scanned)
	return append(manifestEndpoints, scanned...)
}

func addHostRoots(endpoint string, seen map[string]struct{}, add func(string)) {
	host := hostFromEndpoint(endpoint)
	if host == "" {
		return
	}
	root := "https://" + host + "/"
	if _, ok := seen[root]; !ok {
		add(root)
	}
}

func hostFromEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	if m := schemeHostRe.FindStringSubmatch(endpoint); len(m) == 2 {
		return strings.ToLower(m[1])
	}
	if isHTTPURL(endpoint) {
		return ""
	}
	return ""
}

func normalizeEndpoint(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimRight(value, `.,;)]}"'`)
	return value
}

func isHTTPURL(value string) bool {
	lower := strings.ToLower(value)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func isExternalHTTPURL(value string) bool {
	if !isHTTPURL(value) {
		return false
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "localhost") {
		return false
	}
	if strings.Contains(lower, "127.0.0.1") {
		return false
	}
	if strings.Contains(lower, "0.0.0.0") {
		return false
	}
	if strings.Contains(lower, "/_plugins/") {
		return false
	}
	return true
}

func shouldScanNetworkFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js", ".mjs", ".json", ".wasm", ".ts", ".go":
		return true
	default:
		return false
	}
}

func extractURLsFromBytes(data []byte) []string {
	return extractURLsFromText(string(data))
}

func extractURLsFromText(text string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(endpoint string) {
		endpoint = normalizeEndpoint(endpoint)
		if endpoint == "" || !isExternalHTTPURL(endpoint) {
			return
		}
		if _, ok := seen[endpoint]; ok {
			return
		}
		seen[endpoint] = struct{}{}
		out = append(out, endpoint)
	}

	for _, match := range urlInTextRe.FindAllString(text, -1) {
		add(match)
	}
	return out
}

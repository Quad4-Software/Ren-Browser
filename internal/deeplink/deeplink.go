// SPDX-License-Identifier: MIT
package deeplink

import (
	"net/url"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	SchemeRenBrowser = "renbrowser"
	SchemeRNS        = "rns"
	EventName        = "app:deeplink"
)

var (
	mu      sync.Mutex
	pending string
	handler func(string)
)

// SetHandler registers the callback invoked when a deeplink is accepted.
// The handler receives the unwrapped internal navigation URL.
func SetHandler(h func(string)) {
	mu.Lock()
	handler = h
	mu.Unlock()
}

// HandleIncoming unwraps an OS-provided URL, queues it, and notifies the handler.
// Returns the unwrapped target when accepted.
func HandleIncoming(raw string) (string, bool) {
	target, ok := Unwrap(raw)
	if !ok {
		return "", false
	}
	Enqueue(target)
	mu.Lock()
	h := handler
	mu.Unlock()
	if h != nil {
		h(target)
	}
	return target, true
}

// Enqueue stores the latest pending deeplink target (overwrites older ones).
func Enqueue(target string) {
	target = strings.TrimSpace(target)
	if target == "" {
		return
	}
	mu.Lock()
	pending = target
	mu.Unlock()
}

// TakePending returns and clears the pending deeplink target.
func TakePending() string {
	mu.Lock()
	defer mu.Unlock()
	out := pending
	pending = ""
	return out
}

// PeekPending returns the pending deeplink without clearing it.
func PeekPending() string {
	mu.Lock()
	defer mu.Unlock()
	return pending
}

// ClearPending drops any queued deeplink.
func ClearPending() {
	mu.Lock()
	pending = ""
	mu.Unlock()
}

// Unwrap converts an OS launch URL into an internal navigation target.
// Accepted forms include renbrowser: wrappers, rns:// mesh URLs, built-in
// schemes, and bare mesh destinations. External web URLs are rejected.
func Unwrap(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	if containsNUL(raw) || !utf8.ValidString(raw) {
		return "", false
	}

	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "renbrowser:"):
		return unwrapRenBrowser(raw)
	case strings.HasPrefix(lower, "rns://"):
		return normalizeInternal(raw)
	case isBuiltinScheme(lower):
		return normalizeBuiltin(raw), true
	case strings.HasPrefix(lower, "document:"):
		return normalizeDocument(raw), true
	case looksLikeMeshURL(raw):
		return strings.TrimSpace(raw), true
	case isBareNodeHash(raw):
		return strings.ToLower(strings.TrimSpace(raw)) + ":/page/index.mu", true
	default:
		return "", false
	}
}

func unwrapRenBrowser(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	rest := trimmed[len("renbrowser"):]
	if strings.HasPrefix(rest, ":") {
		rest = rest[1:]
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", false
	}

	// renbrowser:about / renbrowser:license: / renbrowser:docs:?page=...
	if !strings.HasPrefix(rest, "//") {
		return unwrapRenBrowserOpaque(rest)
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", false
	}

	host := strings.ToLower(strings.TrimSpace(parsed.Host))
	path := parsed.EscapedPath()
	query := parsed.RawQuery

	// renbrowser://open?url=<encoded>
	if host == "open" || (host == "" && strings.Trim(path, "/") == "open") {
		return unwrapOpenQuery(parsed.Query())
	}

	// renbrowser://about  renbrowser://license  etc.
	if host != "" && (path == "" || path == "/") && query == "" {
		if isBuiltinName(host) {
			return host + ":", true
		}
		if isBareNodeHash(host) {
			return strings.ToLower(host) + ":/page/index.mu", true
		}
	}

	// renbrowser://<32hex>/page/... → mesh URL
	if isBareNodeHash(host) {
		meshPath := path
		if meshPath == "" || meshPath == "/" {
			meshPath = "/page/index.mu"
		}
		target := strings.ToLower(host) + ":" + meshPath
		if query != "" {
			target += "?" + query
		}
		if parsed.Fragment != "" {
			target += "#" + parsed.Fragment
		}
		return target, true
	}

	// renbrowser://rns/<hash>/page/... → rns://...
	if host == "rns" {
		rnsPath := strings.TrimPrefix(path, "/")
		if rnsPath == "" {
			return "", false
		}
		target := "rns://" + rnsPath
		if query != "" {
			target += "?" + query
		}
		return normalizeInternal(target)
	}

	// renbrowser://docs?page=... (host=docs, query present)
	if isBuiltinName(host) {
		target := host + ":"
		if query != "" {
			target += "?" + query
		}
		return target, true
	}

	// Fallback: treat opaque path after scheme as an internal URL.
	candidate := strings.TrimPrefix(rest, "//")
	return unwrapRenBrowserOpaque(candidate)
}

func unwrapRenBrowserOpaque(rest string) (string, bool) {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", false
	}
	lower := strings.ToLower(rest)

	if strings.HasPrefix(lower, "open?") || strings.HasPrefix(lower, "open/?") {
		q := rest
		if i := strings.Index(q, "?"); i >= 0 {
			q = q[i+1:]
		}
		values, err := url.ParseQuery(q)
		if err != nil {
			return "", false
		}
		return unwrapOpenQuery(values)
	}

	if strings.HasPrefix(lower, "url=") {
		values, err := url.ParseQuery(rest)
		if err != nil {
			return "", false
		}
		return unwrapOpenQuery(values)
	}

	if isBuiltinScheme(lower) || strings.HasPrefix(lower, "document:") || strings.HasPrefix(lower, "rns://") || looksLikeMeshURL(rest) {
		return Unwrap(rest)
	}

	if isBuiltinName(lower) || isBuiltinName(strings.TrimSuffix(lower, ":")) {
		return normalizeBuiltin(rest), true
	}

	if isBareNodeHash(rest) {
		return strings.ToLower(rest) + ":/page/index.mu", true
	}

	return "", false
}

func unwrapOpenQuery(values url.Values) (string, bool) {
	if values == nil {
		return "", false
	}
	encoded := strings.TrimSpace(values.Get("url"))
	if encoded == "" {
		encoded = strings.TrimSpace(values.Get("u"))
	}
	if encoded == "" {
		return "", false
	}
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		decoded = encoded
	}
	// Prevent recursive renbrowser wrappers from nesting forever.
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(decoded)), "renbrowser:") {
		inner, ok := Unwrap(decoded)
		return inner, ok
	}
	return Unwrap(decoded)
}

func normalizeInternal(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "http:") || strings.HasPrefix(lower, "https:") ||
		strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") ||
		strings.HasPrefix(lower, "file:") || strings.HasPrefix(lower, "blob:") {
		return "", false
	}
	if strings.HasPrefix(lower, "rns://") {
		return raw, true
	}
	if looksLikeMeshURL(raw) {
		return raw, true
	}
	return "", false
}

func normalizeBuiltin(raw string) string {
	raw = strings.TrimSpace(raw)
	lower := strings.ToLower(raw)
	for _, name := range builtinNames() {
		if lower == name || lower == name+":" {
			return name + ":"
		}
		prefix := name + "?"
		if strings.HasPrefix(lower, prefix) {
			return name + ":?" + raw[len(prefix):]
		}
		prefix = name + ":?"
		if strings.HasPrefix(lower, prefix) {
			return name + ":?" + raw[len(prefix):]
		}
		if strings.HasPrefix(lower, name+":") {
			return name + raw[len(name):]
		}
	}
	return raw
}

func normalizeDocument(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "?") {
		return raw
	}
	rest := raw[len("document:"):]
	if !strings.HasPrefix(rest, "/") {
		rest = "/" + rest
	}
	return "document:" + rest
}

func isBuiltinScheme(lower string) bool {
	for _, name := range builtinNames() {
		if lower == name || lower == name+":" || strings.HasPrefix(lower, name+":") || strings.HasPrefix(lower, name+"?") {
			return true
		}
	}
	return false
}

func isBuiltinName(name string) bool {
	name = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(name, ":")))
	for _, b := range builtinNames() {
		if name == b {
			return true
		}
	}
	return false
}

func builtinNames() []string {
	return []string{"about", "license", "editor", "config", "settings", "docs", "hello"}
}

func looksLikeMeshURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, ":/") {
		return false
	}
	hash, _, ok := strings.Cut(raw, ":/")
	if !ok {
		return false
	}
	return isBareNodeHash(hash)
}

func isBareNodeHash(raw string) bool {
	raw = strings.TrimSpace(raw)
	if len(raw) != 32 {
		return false
	}
	for _, c := range raw {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

func containsNUL(s string) bool {
	return strings.IndexByte(s, 0) >= 0
}

// ExtractFromArgs returns the first argv entry that unwraps to a deeplink target.
func ExtractFromArgs(args []string) (string, bool) {
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" || strings.HasPrefix(arg, "-") {
			continue
		}
		if target, ok := Unwrap(arg); ok {
			return target, true
		}
	}
	return "", false
}

// SPDX-License-Identifier: MIT
package nomadnet

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrEmptyURL    = errors.New("empty url")
	ErrInvalidNode = errors.New("invalid node hash")
)

type PageURL struct {
	NodeHash string
	Path     string
	Request  RequestData
}

func ParseURL(raw string) (PageURL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return PageURL{}, ErrEmptyURL
	}

	if strings.HasPrefix(raw, "rns://") {
		return parseRNSURL(raw)
	}
	if strings.Contains(raw, ":/") {
		return parseMeshURL(raw)
	}
	return PageURL{}, ErrInvalidNode
}

func parseRNSURL(raw string) (PageURL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return PageURL{}, err
	}
	node := strings.Trim(u.Host, "/")
	path := u.Path
	if path == "" {
		path = "/page/index.mu"
	}
	fields := parseQueryFields(u.RawQuery)
	node = normalizeHash(node)
	if node == "" {
		return PageURL{}, ErrInvalidNode
	}
	return PageURL{NodeHash: node, Path: normalizePath(path), Request: parseRequestPairs(fields)}, nil
}

func parseMeshURL(raw string) (PageURL, error) {
	parts := strings.SplitN(raw, ":/", 2)
	if len(parts) != 2 {
		return PageURL{}, ErrInvalidNode
	}
	node := normalizeHash(parts[0])
	if node == "" {
		return PageURL{}, ErrInvalidNode
	}
	rest := parts[1]
	path := rest
	var fields map[string]string
	if qIdx := strings.Index(rest, "?"); qIdx >= 0 {
		path = rest[:qIdx]
		fields = parseFieldSuffix(rest[qIdx+1:])
	} else if idx := strings.Index(rest, "`"); idx >= 0 {
		path = rest[:idx]
		fields = parseBacktickFields(rest[idx+1:])
	}
	return PageURL{NodeHash: node, Path: normalizePath(path), Request: parseRequestPairs(fields)}, nil
}

func parseFieldSuffix(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	if idx := strings.Index(raw, "`"); idx >= 0 {
		out := parseQueryFields(raw[:idx])
		for k, v := range parseBacktickFields(raw[idx+1:]) {
			if out == nil {
				out = make(map[string]string, 4)
			}
			out[k] = v
		}
		return out
	}
	return parseQueryFields(raw)
}

func parseQueryFields(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	var out map[string]string
	for part := range strings.SplitSeq(raw, "&") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			k = part
			v = ""
		}
		key, err := url.QueryUnescape(k)
		if err != nil {
			key = k
		}
		val, err := url.QueryUnescape(v)
		if err != nil {
			val = v
		}
		if out == nil {
			out = make(map[string]string, 4)
		}
		out[key] = val
	}
	return out
}

func parseBacktickFields(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	var out map[string]string
	for part := range strings.SplitSeq(raw, "|") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		if ok {
			if out == nil {
				out = make(map[string]string, 4)
			}
			out[k] = v
		}
	}
	return out
}

func normalizeHash(hash string) string {
	return strings.ToLower(strings.TrimSpace(hash))
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/page/index.mu"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasPrefix(path, "/page/") || strings.HasPrefix(path, "/file/") {
		return path
	}
	if strings.HasSuffix(path, ".mu") || strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".md") {
		return "/page/" + strings.TrimPrefix(path, "/")
	}
	return "/page/" + strings.TrimPrefix(path, "/")
}

func FormatURL(nodeHash, path string) string {
	return normalizeHash(nodeHash) + ":" + normalizePath(path)
}

func FormatURLWithRequest(nodeHash, path string, req RequestData) string {
	base := FormatURL(nodeHash, path)
	pairs := formatRequestPairs(req)
	if len(pairs) == 0 {
		return base
	}
	return base + "`" + strings.Join(pairs, "|")
}

func FormatURLWithFields(nodeHash, path string, fields map[string]string) string {
	return FormatURLWithRequest(nodeHash, path, RequestData{Vars: fields})
}

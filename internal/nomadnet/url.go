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
	node = normalizeHash(node)
	if node == "" {
		return PageURL{}, ErrInvalidNode
	}
	return PageURL{NodeHash: node, Path: normalizePath(path), Request: parseQueryRequest(u.RawQuery)}, nil
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
	var req RequestData
	if before, after, ok := strings.Cut(rest, "?"); ok {
		path = before
		req = parseFieldSuffixRequest(after)
	} else if before, after, ok := strings.Cut(rest, "`"); ok {
		path = before
		req = parseRequestPairs(parseBacktickFields(after))
	}
	return PageURL{NodeHash: node, Path: normalizePath(path), Request: req}, nil
}

func parseFieldSuffixRequest(raw string) RequestData {
	if raw == "" {
		return RequestData{}
	}
	if before, after, ok := strings.Cut(raw, "`"); ok {
		req := parseQueryRequest(before)
		for k, v := range parseBacktickFields(after) {
			req = mergeRequestPair(req, k, v)
		}
		return req
	}
	return parseQueryRequest(raw)
}

func parseQueryRequest(raw string) RequestData {
	if raw == "" {
		return RequestData{}
	}
	n := strings.Count(raw, "&") + 1
	req := RequestData{Vars: make(map[string]string, n)}
	for part := range strings.SplitSeq(raw, "&") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			k = part
			v = ""
		}
		key := queryUnescape(k)
		val := queryUnescape(v)
		if key == "" {
			continue
		}
		kind, name := classifyRequestPair(key)
		if name == "" {
			continue
		}
		if kind == "field" {
			if req.Fields == nil {
				req.Fields = make(map[string]string, 1)
			}
			req.Fields[name] = val
			continue
		}
		req.Vars[name] = val
	}
	if len(req.Vars) == 0 {
		req.Vars = nil
	}
	return req
}

func queryUnescape(s string) string {
	if !queryNeedsUnescape(s) {
		return s
	}
	out, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return out
}

func queryNeedsUnescape(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '%' || c == '+' {
			return true
		}
	}
	return false
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
				out = make(map[string]string, strings.Count(raw, "|")+1)
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

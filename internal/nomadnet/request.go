// SPDX-License-Identifier: MIT
package nomadnet

import (
	"sort"
	"strings"
)

type RequestData struct {
	Vars   map[string]string
	Fields map[string]string
}

func (r RequestData) Empty() bool {
	return len(r.Vars) == 0 && len(r.Fields) == 0
}

// buildRequestData assembles the var_/field_ prefixed map that Nomad Network
// nodes expect as the request payload. It must be handed to link.Request as
// a plain map (or nil), never pre-encoded to bytes: Nomad Network only reads
// request parameters when the unpacked wire payload is a dict
// (isinstance(data, dict) in nomadnet/Node.py). Pre-msgpacking it here would
// ship a bytes blob instead and silently drop every var/field parameter.
func buildRequestData(req RequestData) map[string]string {
	if req.Empty() {
		return nil
	}
	out := make(map[string]string, len(req.Vars)+len(req.Fields))
	for k, v := range req.Vars {
		out["var_"+k] = v
	}
	for k, v := range req.Fields {
		out["field_"+k] = v
	}
	return out
}

const fieldKeyPrefix = "field."

func classifyRequestPair(key string) (kind, name string) {
	if after, ok := strings.CutPrefix(key, fieldKeyPrefix); ok {
		return "field", after
	}
	return "var", key
}

func mergeRequestPair(req RequestData, key, value string) RequestData {
	kind, name := classifyRequestPair(key)
	if name == "" {
		return req
	}
	switch kind {
	case "field":
		if req.Fields == nil {
			req.Fields = make(map[string]string, 1)
		}
		req.Fields[name] = value
	default:
		if req.Vars == nil {
			req.Vars = make(map[string]string, 1)
		}
		req.Vars[name] = value
	}
	return req
}

func parseRequestPairs(pairs map[string]string) RequestData {
	if len(pairs) == 0 {
		return RequestData{}
	}
	var req RequestData
	for k, v := range pairs {
		req = mergeRequestPair(req, k, v)
	}
	return req
}

func formatRequestPairs(req RequestData) []string {
	if req.Empty() {
		return nil
	}
	pairs := make([]string, 0, len(req.Vars)+len(req.Fields))
	for k, v := range req.Vars {
		pairs = append(pairs, k+"="+v)
	}
	for k, v := range req.Fields {
		pairs = append(pairs, fieldKeyPrefix+k+"="+v)
	}
	sort.Strings(pairs)
	return pairs
}

func (r RequestData) CacheKeySuffix() string {
	pairs := formatRequestPairs(r)
	if len(pairs) == 0 {
		return ""
	}
	return "`" + strings.Join(pairs, "|")
}

// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"maps"
	"strings"
)

// FieldInput is a normalized HTML control snapshot (type, name, value, checked).
// Type names are matched case-insensitively. Checkbox and radio inputs use Checked.
type FieldInput struct {
	Type    string
	Name    string
	Value   string
	Checked bool
}

// RequestPayload is the resolved destination, submitted fields, and backtick
// request variables from a Micron-style link destination string.
type RequestPayload struct {
	Destination string            `json:"destination"`
	Fields      map[string]string `json:"fields"`
	RequestVars map[string]string `json:"request_vars"`
}

// CollectFormFields converts HTML input snapshots into a name to value map.
// Checkboxes that share a name and are all checked get their values joined with commas.
// For radios, the last checked input for a given name wins.
func CollectFormFields(inputs []FieldInput) map[string]string {
	out := map[string]string{}
	for _, in := range inputs {
		if in.Name == "" {
			continue
		}
		t := strings.ToLower(strings.TrimSpace(in.Type))
		switch t {
		case "checkbox":
			if !in.Checked {
				continue
			}
			if prev, ok := out[in.Name]; ok && prev != "" {
				out[in.Name] = prev + "," + in.Value
			} else {
				out[in.Name] = in.Value
			}
		case "radio":
			if in.Checked {
				out[in.Name] = in.Value
			}
		default:
			out[in.Name] = in.Value
		}
	}
	return out
}

// BuildRequestPayload splits destination on "`" into a base path and optional
// k=v|… request variables, then selects fields from allFields using fieldsSpec.
// Use "*" to copy every entry from allFields. The map allFields is not modified.
func BuildRequestPayload(allFields map[string]string, destination, fieldsSpec string) RequestPayload {
	dest, reqVars := splitDestinationVars(destination)
	selected := map[string]string{}
	if fieldsSpec == "*" {
		maps.Copy(selected, allFields)
	} else {
		for _, name := range splitFieldList(fieldsSpec) {
			if v, ok := allFields[name]; ok {
				selected[name] = v
			}
		}
	}
	return RequestPayload{
		Destination: dest,
		Fields:      selected,
		RequestVars: reqVars,
	}
}

func splitFieldList(fieldsSpec string) []string {
	if strings.TrimSpace(fieldsSpec) == "" {
		return nil
	}
	parts := strings.Split(fieldsSpec, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func splitDestinationVars(destination string) (string, map[string]string) {
	destination = strings.TrimSpace(destination)
	vars := map[string]string{}
	if destination == "" {
		return "", vars
	}
	before, after, ok := strings.Cut(destination, "`")
	if !ok {
		return destination, vars
	}
	base := before
	raw := after
	if raw == "" {
		return base, vars
	}
	for pair := range strings.SplitSeq(raw, "|") {
		if pair == "" {
			continue
		}
		eq := strings.IndexByte(pair, '=')
		if eq <= 0 {
			continue
		}
		k := strings.TrimSpace(pair[:eq])
		v := strings.TrimSpace(pair[eq+1:])
		if k == "" {
			continue
		}
		vars[k] = v
	}
	return base, vars
}

// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "encoding/json"

// DocumentJSON returns a JSON encoding of the Document IR (blocks, spans, diagnostics).
func (doc *Document) DocumentJSON() ([]byte, error) {
	if doc == nil {
		return []byte("null"), nil
	}
	return json.Marshal(doc)
}

// DiagnosticsJSON returns JSON for a diagnostic slice.
func DiagnosticsJSON(diags []Diagnostic) ([]byte, error) {
	if diags == nil {
		diags = []Diagnostic{}
	}
	return json.Marshal(diags)
}

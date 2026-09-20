// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

// Span is a half-open byte range [Start, End) into the original markup.
type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// BlockKind classifies a top-level document block.
type BlockKind int

const (
	BlockBlank BlockKind = iota
	BlockParagraph
	BlockHeading
	BlockHR
	BlockDivider
	BlockPartial
)

// Document is the public Micron IR produced by Parse.
type Document struct {
	// Source is the original markup when Parse retained it (always set by Parse).
	Source string `json:"-"`
	// Colors are page-level #!fg / #!bg values from leading directives.
	Colors PageColors `json:"colors"`
	// DefaultFG and DefaultBG are the resolved palette defaults after headers and theme.
	DefaultFG string `json:"default_fg"`
	DefaultBG string `json:"default_bg"`
	// Blocks is the document body in source order.
	Blocks []Block `json:"blocks"`
	// Diagnostics is non-nil when ParseWithDiagnostics collected issues.
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

// Block is one structural unit (line or heading/partial/hr).
type Block struct {
	Kind BlockKind `json:"kind"`
	Span Span      `json:"span"`
	// SourceLine is the 1-based source line for data-mu-line (set by Parse).
	SourceLine   int      `json:"source_line,omitempty"`
	Depth        int      `json:"depth"`
	Align        string   `json:"align,omitempty"`
	LineBG       string   `json:"line_bg,omitempty"`
	HeadingStyle Style    `json:"heading_style"`
	DividerRune  rune     `json:"divider_rune,omitempty"`
	Literal      bool     `json:"literal,omitempty"`
	Inlines      []Inline `json:"inlines,omitempty"`
	Partial      *Partial `json:"partial,omitempty"`
}

// Inline is a styled text run or an embedded widget on a line.
type Inline struct {
	Span    Span     `json:"span"`
	Style   Style    `json:"style"`
	Text    string   `json:"text,omitempty"`
	Field   *Field   `json:"field,omitempty"`
	Link    *Link    `json:"link,omitempty"`
	Partial *Partial `json:"partial,omitempty"`
}

// Severity ranks a diagnostic.
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

// Diagnostic is an authoring or lint finding with an optional source span.
type Diagnostic struct {
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Span     Span     `json:"span"`
}

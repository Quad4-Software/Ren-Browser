// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

// Parser configures Micron-to-HTML conversion. A zero Parser is usable.
// Set DarkTheme and ForceMonospace for light/dark defaults and monospace layout.
// Parser values are safe for concurrent use by multiple goroutines.
type Parser struct {
	// DarkTheme selects default palette when the document does not set #!fg / #!bg.
	DarkTheme bool
	// ForceMonospace wraps segments in monospace span markup and affects how
	// link labels are escaped.
	ForceMonospace bool
}

// Formatting tracks inline text styling inside backtick formatting mode.
type Formatting struct {
	Bold      bool
	Underline bool
	Italic    bool
}

// Style is a resolved foreground, background, and font style.
type Style struct {
	FG        string
	BG        string
	Bold      bool
	Underline bool
	Italic    bool
}

// State holds parser state across lines.
type State struct {
	Literal        bool
	TableMode      bool
	TableLines     []string
	TableOptsAlign string
	TableOptsMaxW  int
	Depth          int
	FGColor        string
	BGColor        string
	Formatting     Formatting
	DefaultAlign   string
	Align          string
	DefaultFG      string
	DefaultBG      string
	styleAttrMap   map[stateStyleKey]string
}

type stateStyleKey struct {
	FG        string
	BG        string
	Bold      bool
	Underline bool
	Italic    bool
}

// FieldKind selects the widget produced by a field span.
type FieldKind int

const (
	FieldText FieldKind = iota
	FieldCheckbox
	FieldRadio
)

// Field is a text field, checkbox, or radio control.
type Field struct {
	Kind       FieldKind
	Name       string
	Value      string
	Label      string
	Width      int
	Masked     bool
	Prechecked bool
	Style      Style
}

// Link is an anchor with optional field submission metadata.
type Link struct {
	URL    string
	Label  string
	Fields []string
	Style  Style
}

// Partial is an asynchronously loaded micron block (placeholder ⧖) with optional
// refresh interval and field metadata, matching micron-parser-js / NomadNet.
type Partial struct {
	URL         string
	Destination string
	Descriptor  string
	FieldsAttr  string
	PartialID   string
	HasRefresh  bool
	Refresh     float64
	Style       Style
}

// linePart is one segment on a line after makeOutput.
type linePart struct {
	style   Style
	text    string
	html    string
	field   *Field
	link    *Link
	partial *Partial
}

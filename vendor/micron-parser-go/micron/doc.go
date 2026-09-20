// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

// Package micron parses Micron markup into a Document IR and renders HTML
// fragments intended for embedding in a host page.
//
// # Authority
//
// NomadNet Python is the primary dialect authority
// (nomadnet/ui/textui/MicronParser.py and Guide.py). micron-parser-js is a
// secondary interop target for web hosts. When Go and JS disagree with Python,
// prefer Python and document the JS divergence.
//
// # HTML and security
//
// ConvertMicronToHTML returns a fragment. User text is escaped and attribute values on
// generated elements use HTML escaping. The host still decides how to mount that fragment
// (innerHTML vs safer DOM APIs), CSP, and how link destinations and partial URLs are fetched.
// Link href and data-* strings follow Micron / NomadNet URL rules (see FormatNomadnetworkURL).
// That is not a general-purpose URL allowlist for arbitrary schemes.
//
// # Document IR
//
// Parse builds a Document with source Spans. RenderHTML turns a Document into HTML.
// ConvertMicronToHTML is Parse followed by RenderHTML. ParseWithDiagnostics and Lint
// collect authoring findings without changing silent convert behavior.
//
// # Concurrency
//
// Parser holds only DarkTheme and ForceMonospace.
// There is no per-conversion mutable state. The same Parser may be shared across goroutines.
//
// # Interop
//
// Tests check behavioral parity with micron-parser-js by comparing structural
// signatures of HTML output rather than byte-identical strings.
// The JS script lives under micron/testdata/. Python oracle tests run when
// python3 and optionally nomadnet are available.
package micron

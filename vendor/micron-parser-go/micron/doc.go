// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

// Package micron parses Micron markup and renders HTML fragments intended for
// embedding in a host page.
//
// # HTML and security
//
// ConvertMicronToHTML returns a fragment. User text is escaped and attribute values on
// generated elements use HTML escaping. The host still decides how to mount that fragment
// (innerHTML vs safer DOM APIs), CSP, and how link destinations and partial URLs are fetched.
// Link href and data-* strings follow Micron / NomadNet URL rules (see FormatNomadnetworkURL).
// That is not a general-purpose URL allowlist for arbitrary schemes.
//
// # Concurrency
//
// Parser holds only DarkTheme and ForceMonospace.
// There is no per-conversion mutable state. The same Parser may be shared across goroutines.
//
// # Reference implementation
//
// Tests check behavioral parity with the JavaScript reference by comparing structural
// signatures of HTML output rather than byte-identical strings.
// The reference script lives under micron/testdata/.
package micron

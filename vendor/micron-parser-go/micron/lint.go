// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"unicode"
)

// Lint analyzes markup and returns diagnostics without requiring a full Document walk beyond Parse.
func (p *Parser) Lint(markup string) []Diagnostic {
	_, diags := p.ParseWithDiagnostics(markup)
	return diags
}

func lintHeaderTags(markup string, diags *[]Diagnostic) {
	off := 0
	for off <= len(markup) {
		rel := strings.IndexByte(markup[off:], '\n')
		end := len(markup)
		next := len(markup) + 1
		if rel >= 0 {
			end = off + rel
			next = end + 1
		}
		line := markup[off:end]
		t := strings.TrimSpace(line)
		if t == "" {
			off = next
			continue
		}
		if !strings.HasPrefix(t, "#!") {
			break
		}
		span := Span{Start: off, End: end}
		if strings.HasPrefix(t, "#!fg=") {
			c := strings.TrimSpace(t[5:])
			if c != "" && len(c) != 3 && len(c) != 6 {
				*diags = append(*diags, Diagnostic{
					Severity: SeverityWarning,
					Code:     "header.fg_invalid",
					Message:  "#!fg= expects 3 or 6 hex digits",
					Span:     span,
				})
			} else if c != "" && !isHexDigits(c) {
				*diags = append(*diags, Diagnostic{
					Severity: SeverityWarning,
					Code:     "header.fg_invalid",
					Message:  "#!fg= color is not hexadecimal",
					Span:     span,
				})
			}
		} else if strings.HasPrefix(t, "#!bg=") {
			c := strings.TrimSpace(t[5:])
			if c != "" && len(c) != 3 && len(c) != 6 {
				*diags = append(*diags, Diagnostic{
					Severity: SeverityWarning,
					Code:     "header.bg_invalid",
					Message:  "#!bg= expects 3 or 6 hex digits",
					Span:     span,
				})
			} else if c != "" && !isHexDigits(c) {
				*diags = append(*diags, Diagnostic{
					Severity: SeverityWarning,
					Code:     "header.bg_invalid",
					Message:  "#!bg= color is not hexadecimal",
					Span:     span,
				})
			}
		}
		off = next
	}
	lintDocumentExtras(markup, diags)
}

func isHexDigits(s string) bool {
	for _, r := range s {
		if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
			return false
		}
	}
	return true
}

func lintDocumentExtras(markup string, diags *[]Diagnostic) {
	off := 0
	for off <= len(markup) {
		rel := strings.IndexByte(markup[off:], '\n')
		end := len(markup)
		next := len(markup) + 1
		if rel >= 0 {
			end = off + rel
			next = end + 1
		}
		line := markup[off:end]
		span := Span{Start: off, End: end}
		if strings.Contains(line, "`[") && !strings.Contains(line, "]") {
			*diags = append(*diags, Diagnostic{
				Severity: SeverityWarning,
				Code:     "link.unclosed",
				Message:  "link opener `[ without closing ] on this line",
				Span:     span,
			})
		}
		if strings.Contains(line, "`<") && !strings.Contains(line, ">") {
			*diags = append(*diags, Diagnostic{
				Severity: SeverityWarning,
				Code:     "field.unclosed",
				Message:  "field opener `< without closing > on this line",
				Span:     span,
			})
		}
		trim := trimASCIISpaces(line)
		if strings.HasPrefix(trim, "`[") && strings.HasSuffix(trim, "]") {
			inner := trim[2 : len(trim)-1]
			if _, after, ok := strings.Cut(inner, "`"); ok {
				url := after
				if before, a2, ok2 := strings.Cut(after, "`"); ok2 {
					url = before
					_ = a2
				}
				if url != "" {
					low := strings.ToLower(url)
					if strings.HasPrefix(low, "javascript:") || strings.HasPrefix(low, "data:") {
						*diags = append(*diags, Diagnostic{
							Severity: SeverityWarning,
							Code:     "url.dangerous_scheme",
							Message:  "link URL uses a neutralized or dangerous scheme",
							Span:     span,
						})
					}
				}
			}
		}
		if fg, bg, ok := pageColorsFromLine(line); ok {
			if contrastSuspect(fg, bg) {
				*diags = append(*diags, Diagnostic{
					Severity: SeverityInfo,
					Code:     "color.low_contrast",
					Message:  "page fg/bg may have low contrast",
					Span:     span,
				})
			}
		}
		off = next
	}
}

func pageColorsFromLine(line string) (fg, bg string, ok bool) {
	t := strings.TrimSpace(line)
	if strings.HasPrefix(t, "#!fg=") {
		return strings.TrimSpace(t[5:]), "", true
	}
	if strings.HasPrefix(t, "#!bg=") {
		return "", strings.TrimSpace(t[5:]), true
	}
	return "", "", false
}

func contrastSuspect(fg, bg string) bool {
	if fg == "" || bg == "" {
		return false
	}
	if len(fg) == 3 && len(bg) == 3 && fg == bg {
		return true
	}
	if len(fg) == 6 && len(bg) == 6 && fg == bg {
		return true
	}
	return false
}

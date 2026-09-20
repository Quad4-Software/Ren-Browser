// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import "strings"

// Builder constructs Micron markup safely without string-soup concatenation errors.
type Builder struct {
	b strings.Builder
}

// String returns the accumulated Micron source.
func (b *Builder) String() string { return b.b.String() }

// Reset clears the builder.
func (b *Builder) Reset() { b.b.Reset() }

// Text appends a plain text line (no trailing newline added for mid-line use).
func (b *Builder) Text(s string) *Builder {
	b.b.WriteString(s)
	return b
}

// Line appends s and a newline.
func (b *Builder) Line(s string) *Builder {
	b.b.WriteString(s)
	b.b.WriteByte('\n')
	return b
}

// NL appends a blank line.
func (b *Builder) NL() *Builder {
	b.b.WriteByte('\n')
	return b
}

// Heading appends a section heading at the given depth (clamped 1..16).
func (b *Builder) Heading(depth int, title string) *Builder {
	if depth < 1 {
		depth = 1
	}
	if depth > maxSectionDepth {
		depth = maxSectionDepth
	}
	for range depth {
		b.b.WriteByte('>')
	}
	if title != "" {
		b.b.WriteByte(' ')
		b.b.WriteString(title)
	}
	b.b.WriteByte('\n')
	return b
}

// Bold wraps s in bold toggles.
func (b *Builder) Bold(s string) *Builder {
	b.b.WriteString("`!")
	b.b.WriteString(s)
	b.b.WriteString("`!")
	return b
}

// Italic wraps s in italic toggles.
func (b *Builder) Italic(s string) *Builder {
	b.b.WriteString("`*")
	b.b.WriteString(s)
	b.b.WriteString("`*")
	return b
}

// Link appends a Micron link. label may be empty (URL is shown).
func (b *Builder) Link(label, url string, fields ...string) *Builder {
	b.b.WriteString("`[")
	b.b.WriteString(label)
	b.b.WriteByte('`')
	b.b.WriteString(url)
	if len(fields) > 0 {
		b.b.WriteByte('`')
		b.b.WriteString(strings.Join(fields, "|"))
	}
	b.b.WriteByte(']')
	return b
}

// FieldText appends a text field.
func (b *Builder) FieldText(name, value string, width int) *Builder {
	b.b.WriteString("`<")
	if width > 0 {
		b.b.WriteString(itoa(width))
		b.b.WriteByte('|')
	}
	b.b.WriteString(name)
	b.b.WriteByte('`')
	b.b.WriteString(value)
	b.b.WriteByte('>')
	return b
}

// FieldCheckbox appends a checkbox field.
func (b *Builder) FieldCheckbox(name, value, label string, checked bool) *Builder {
	b.b.WriteString("`<?|")
	b.b.WriteString(name)
	b.b.WriteByte('|')
	b.b.WriteString(value)
	if checked {
		b.b.WriteString("|*")
	}
	b.b.WriteByte('`')
	b.b.WriteString(label)
	b.b.WriteByte('>')
	return b
}

// Partial appends an async partial line.
func (b *Builder) Partial(url string, refresh float64, fields string) *Builder {
	b.b.WriteString("`{")
	b.b.WriteString(url)
	if refresh >= 1 {
		b.b.WriteByte('`')
		b.b.WriteString(formatPartialRefresh(refresh))
	}
	if fields != "" {
		if refresh < 1 {
			b.b.WriteString("`0")
		}
		b.b.WriteByte('`')
		b.b.WriteString(fields)
	}
	b.b.WriteByte('}')
	b.b.WriteByte('\n')
	return b
}

// HR appends a horizontal rule.
func (b *Builder) HR() *Builder {
	b.b.WriteString("-\n")
	return b
}

// FG sets foreground color for following text (3 or 6 hex).
func (b *Builder) FG(hex string) *Builder {
	b.b.WriteString("`F")
	b.b.WriteString(hex)
	return b
}

// HeaderColors prepends page #!fg / #!bg directives.
func (b *Builder) HeaderColors(fg, bg string) *Builder {
	if fg != "" {
		b.b.WriteString("#!fg=")
		b.b.WriteString(fg)
		b.b.WriteByte('\n')
	}
	if bg != "" {
		b.b.WriteString("#!bg=")
		b.b.WriteString(bg)
		b.b.WriteByte('\n')
	}
	return b
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [16]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

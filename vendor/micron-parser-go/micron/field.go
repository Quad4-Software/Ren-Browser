// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strconv"
	"strings"
)

func (p *Parser) parseField(line string, start int, s *State) (skip int, f *Field) {
	if start < 0 || start >= len(line) {
		return 0, nil
	}
	fieldStart := start + 1
	bt := strings.IndexByte(line[fieldStart:], '`')
	if bt < 0 {
		return 0, nil
	}
	bt += fieldStart
	fieldContent := line[fieldStart:bt]
	masked := false
	width := 24
	kind := FieldText
	name := fieldContent
	value := ""
	prechecked := false

	if before, after, ok := strings.Cut(fieldContent, "|"); ok {
		flags := before
		rest := after
		name = rest
		value = ""
		if next := strings.IndexByte(rest, '|'); next >= 0 {
			name = rest[:next]
			rest = rest[next+1:]
			value = rest
			if before, after, ok := strings.Cut(rest, "|"); ok {
				value = before
				prechecked = after == "*"
			}
		}
		if strings.Contains(flags, "^") {
			kind = FieldRadio
			flags = stripByte(flags, '^')
		} else if strings.Contains(flags, "?") {
			kind = FieldCheckbox
			flags = stripByte(flags, '?')
		} else if strings.Contains(flags, "!") {
			masked = true
			flags = stripByte(flags, '!')
		}
		if flags != "" {
			if w, err := strconv.Atoi(flags); err == nil {
				if w > 256 {
					w = 256
				}
				if w > 0 {
					width = w
				}
			}
		}
	}

	end := strings.IndexByte(line[bt+1:], '>')
	if end < 0 {
		return 0, nil
	}
	end += bt + 1
	data := line[bt+1 : end]
	st := p.stateToStyle(s)
	sk := end - start + 1

	switch kind {
	case FieldCheckbox, FieldRadio:
		v := value
		if v == "" {
			v = data
		}
		return sk, &Field{
			Kind:       kind,
			Name:       name,
			Value:      v,
			Label:      data,
			Prechecked: prechecked,
			Style:      st,
		}
	default:
		return sk, &Field{
			Kind:   FieldText,
			Name:   name,
			Width:  width,
			Masked: masked,
			Value:  data,
			Style:  st,
		}
	}
}

func stripByte(s string, c byte) string {
	if strings.IndexByte(s, c) < 0 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != c {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

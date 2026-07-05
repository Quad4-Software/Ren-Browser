// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	tableMinColWidth = 3
	defaultTableMaxW = 100
)

var (
	tableStripVisRE1 = regexp.MustCompile("`[FB][0-9a-fA-F]{3}")
	tableStripVisRE2 = regexp.MustCompile("`[FB]T[0-9a-fA-F]{6}")
	tableStripVisRE3 = regexp.MustCompile("`[!*_=]")
	tableStripVisRE4 = regexp.MustCompile("`f`b")
	tableStripVisRE5 = regexp.MustCompile("`f")
	tableStripVisRE6 = regexp.MustCompile("`b")
	tableTruncRE1    = regexp.MustCompile("`[FB][0-9a-fA-F]{3}")
	tableTruncRE2    = regexp.MustCompile("`[!*_]")
	tableTruncRE3    = regexp.MustCompile("`f`b")
)

func stripMicronForVisibleWidth(s string) string {
	s = tableStripVisRE1.ReplaceAllString(s, "")
	s = tableStripVisRE2.ReplaceAllString(s, "")
	s = tableStripVisRE3.ReplaceAllString(s, "")
	s = tableStripVisRE4.ReplaceAllString(s, "")
	s = tableStripVisRE5.ReplaceAllString(s, "")
	s = tableStripVisRE6.ReplaceAllString(s, "")
	return s
}

func visibleDisplayWidth(s string) int {
	return utf8.RuneCountInString(stripMicronForVisibleWidth(s))
}

func escapeTableBackticks(s string) string {
	return strings.ReplaceAll(s, "`", "\\`")
}

func parseTableRow(line string) []string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "|") {
		line = line[1:]
	}
	if strings.HasSuffix(line, "|") {
		line = line[:len(line)-1]
	}
	var cells []string
	var cur strings.Builder
	escaped := false
	for _, r := range line {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}
		switch r {
		case '\\':
			escaped = true
		case '|':
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	cells = append(cells, strings.TrimSpace(cur.String()))
	return cells
}

func parseTableAlignments(line string) []string {
	cells := parseTableRow(line)
	align := make([]string, 0, len(cells))
	for _, cell := range cells {
		c := strings.TrimSpace(cell)
		switch {
		case strings.HasPrefix(c, ":") && strings.HasSuffix(c, ":"):
			align = append(align, "center")
		case strings.HasSuffix(c, ":"):
			align = append(align, "right")
		default:
			align = append(align, "left")
		}
	}
	return align
}

func truncateTableCell(text string, width int) string {
	if visibleDisplayWidth(text) <= width {
		return text
	}
	stripped := text
	stripped = tableTruncRE1.ReplaceAllString(stripped, "")
	stripped = tableTruncRE2.ReplaceAllString(stripped, "")
	stripped = tableTruncRE3.ReplaceAllString(stripped, "")
	if utf8.RuneCountInString(stripped) <= width-1 {
		return text
	}
	rs := []rune(stripped)
	if len(rs) < width {
		return text
	}
	return string(rs[:width-1]) + "…"
}

func padTableCell(text string, width int, align string) string {
	text = truncateTableCell(text, width)
	tw := visibleDisplayWidth(text)
	padding := width - tw
	if padding <= 0 {
		return text
	}
	switch align {
	case "right":
		return strings.Repeat(" ", padding) + text
	case "center":
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	default:
		return text + strings.Repeat(" ", padding)
	}
}

// formatTableRaw converts GitHub-style pipe rows (header, separator, body…) into
// Micron box-drawing lines, matching Reticulum MarkdownToMicron.format_table_raw.
func formatTableRaw(rows []string, blockAlign string, maxWidth int) []string {
	if len(rows) < 2 {
		return nil
	}
	if maxWidth <= 0 {
		maxWidth = defaultTableMaxW
	}

	headerCells := parseTableRow(rows[0])
	if len(headerCells) == 0 {
		return nil
	}
	alignments := parseTableAlignments(rows[1])
	for len(alignments) < len(headerCells) {
		alignments = append(alignments, "left")
	}
	if len(alignments) > len(headerCells) {
		alignments = alignments[:len(headerCells)]
	}

	var dataRows [][]string
	for i := 2; i < len(rows); i++ {
		cells := parseTableRow(rows[i])
		for len(cells) < len(headerCells) {
			cells = append(cells, "")
		}
		if len(cells) > len(headerCells) {
			cells = cells[:len(headerCells)]
		}
		dataRows = append(dataRows, cells)
	}

	numCols := len(headerCells)
	colWidths := make([]int, numCols)
	allRows := [][]string{headerCells}
	for i := range dataRows {
		allRows = append(allRows, dataRows[i])
	}
	for _, row := range allRows {
		for i, cell := range row {
			if i < len(colWidths) {
				w := visibleDisplayWidth(cell)
				if w > colWidths[i] {
					colWidths[i] = w
				}
			}
		}
	}
	for i := range colWidths {
		if colWidths[i] < tableMinColWidth {
			colWidths[i] = tableMinColWidth
		}
	}

	totalWidth := 0
	for _, w := range colWidths {
		totalWidth += w
	}
	totalWidth += numCols*3 + 1
	if totalWidth > maxWidth {
		excess := totalWidth - maxWidth
		type iw struct {
			i int
			w int
		}
		indexed := make([]iw, len(colWidths))
		for i, w := range colWidths {
			indexed[i] = iw{i, w}
		}
		for a := range indexed {
			for b := a + 1; b < len(indexed); b++ {
				if indexed[b].w > indexed[a].w {
					indexed[a], indexed[b] = indexed[b], indexed[a]
				}
			}
		}
		for _, e := range indexed {
			if excess <= 0 {
				break
			}
			red := excess
			maxRed := max(colWidths[e.i]-tableMinColWidth, 0)
			if red > maxRed {
				red = maxRed
			}
			colWidths[e.i] -= red
			excess -= red
		}
	}

	const (
		h  = "─"
		v  = "│"
		tl = "┌"
		tr = "┐"
		bl = "└"
		br = "┘"
		ml = "├"
		mr = "┤"
		tm = "┬"
		bm = "┴"
		mm = "┼"
	)

	var out []string
	if blockAlign == "l" || blockAlign == "c" || blockAlign == "r" {
		out = append(out, "`"+blockAlign)
	}

	var top strings.Builder
	top.WriteString(tl)
	for i, w := range colWidths {
		top.WriteString(strings.Repeat(h, w+2))
		if i < len(colWidths)-1 {
			top.WriteString(tm)
		} else {
			top.WriteString(tr)
		}
	}
	out = append(out, escapeTableBackticks(top.String()))

	var hdr strings.Builder
	hdr.WriteString(v)
	for i, cell := range headerCells {
		pad := padTableCell(cell, colWidths[i], "left")
		hdr.WriteString(" " + pad + " " + v)
	}
	out = append(out, hdr.String())

	var sep strings.Builder
	sep.WriteString(ml)
	for i, w := range colWidths {
		sep.WriteString(strings.Repeat(h, w+2))
		if i < len(colWidths)-1 {
			sep.WriteString(mm)
		} else {
			sep.WriteString(mr)
		}
	}
	out = append(out, escapeTableBackticks(sep.String()))

	for _, row := range dataRows {
		var rowLine strings.Builder
		rowLine.WriteString(v)
		for i, cell := range row {
			al := "left"
			if i < len(alignments) {
				al = alignments[i]
			}
			pad := padTableCell(cell, colWidths[i], al)
			rowLine.WriteString(" " + pad + " " + v)
		}
		out = append(out, rowLine.String())
	}

	var bot strings.Builder
	bot.WriteString(bl)
	for i, w := range colWidths {
		bot.WriteString(strings.Repeat(h, w+2))
		if i < len(colWidths)-1 {
			bot.WriteString(bm)
		} else {
			bot.WriteString(br)
		}
	}
	out = append(out, escapeTableBackticks(bot.String()))

	if blockAlign == "l" || blockAlign == "c" || blockAlign == "r" {
		out = append(out, "`a")
	}
	return out
}

func parseTableFenceOptions(rest string) (align string, maxW int) {
	if len(rest) == 0 {
		return "", 0
	}
	switch rest[0] {
	case 'l', 'c', 'r':
		align = rest[:1]
		rest = rest[1:]
	default:
		align = ""
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return align, 0
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n <= 0 {
		return align, 0
	}
	return align, n
}

func (p *Parser) consumeTableFence(out *strings.Builder, line string, s *State) int {
	if s.TableMode {
		optsAlign := s.TableOptsAlign
		optsMax := s.TableOptsMaxW
		rows := s.TableLines
		s.TableMode = false
		s.TableLines = nil
		s.TableOptsAlign = ""
		s.TableOptsMaxW = 0

		useMax := optsMax
		if useMax <= 0 {
			useMax = defaultTableMaxW
		}
		micronLines := formatTableRaw(rows, optsAlign, useMax)
		if len(micronLines) == 0 {
			return lineOmit
		}
		for _, ml := range micronLines {
			p.parseLineInto(out, ml, s)
		}
		return lineHTML
	}

	rest := line[2:]
	align, maxW := parseTableFenceOptions(rest)
	s.TableMode = true
	if s.TableLines == nil {
		s.TableLines = make([]string, 0, 16)
	} else {
		s.TableLines = s.TableLines[:0]
	}
	s.TableOptsAlign = align
	s.TableOptsMaxW = maxW
	return lineOmit
}

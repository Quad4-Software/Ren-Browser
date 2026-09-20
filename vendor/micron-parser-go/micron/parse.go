// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strings"
	"unicode/utf8"
)

// Parse builds a Document IR from Micron markup. It never returns an error.
// Malformed constructs are omitted or kept as text in NomadNet-compatible fashion.
func (p *Parser) Parse(markup string) *Document {
	doc, _ := p.parseDocument(markup, false)
	return doc
}

// ParseWithDiagnostics builds a Document and a diagnostic list for authoring tools.
func (p *Parser) ParseWithDiagnostics(markup string) (*Document, []Diagnostic) {
	return p.parseDocument(markup, true)
}

func (p *Parser) parseDocument(markup string, collectDiag bool) (*Document, []Diagnostic) {
	pc := ParseHeaderTags(markup)
	plain := plainStyle(p)
	defaultFG := pc.FG
	if defaultFG == "" {
		defaultFG = plain.FG
	}
	defaultBGVal := plain.BG
	if pc.BG != "" {
		defaultBGVal = pc.BG
	}
	s := State{
		Literal:      false,
		Depth:        0,
		FGColor:      defaultFG,
		BGColor:      defaultBGVal,
		DefaultAlign: "start",
		Align:        "start",
		DefaultFG:    defaultFG,
		DefaultBG:    defaultBGVal,
	}
	doc := &Document{
		Source:    markup,
		Colors:    pc,
		DefaultFG: defaultFG,
		DefaultBG: defaultBGVal,
		Blocks:    make([]Block, 0, 32),
	}
	var diags []Diagnostic
	if collectDiag {
		diags = make([]Diagnostic, 0, 8)
		lintHeaderTags(markup, &diags)
	}
	lineNum := 1
	for start := 0; start <= len(markup); {
		nextRel := strings.IndexByte(markup[start:], '\n')
		lineEnd := len(markup)
		next := len(markup) + 1
		if nextRel >= 0 {
			lineEnd = start + nextRel
			next = lineEnd + 1
		}
		line := markup[start:lineEnd]
		lineSpan := Span{Start: start, End: lineEnd}
		p.appendBlocksFromLine(doc, line, lineSpan, lineNum, &s, collectDiag, &diags)
		start = next
		lineNum++
	}
	if collectDiag {
		doc.Diagnostics = diags
	}
	return doc, diags
}

func (p *Parser) appendBlocksFromLine(doc *Document, line string, lineSpan Span, srcLine int, s *State, collectDiag bool, diags *[]Diagnostic) {
	if line != "" {
		if isLiteralToggleLine(line) {
			s.Literal = !s.Literal
			return
		}
		preEscape := false
		if !s.Literal {
			if line[0] == '>' && strings.Contains(line, "`<") {
				k := 0
				for k < len(line) && line[k] == '>' {
					k++
				}
				line = line[k:]
				lineSpan.Start += k
				if line == "" {
					p.appendBlocksFromLine(doc, "", Span{Start: lineSpan.End, End: lineSpan.End}, srcLine, s, collectDiag, diags)
					return
				}
			}
			if line[0] == '\\' {
				line = line[1:]
				lineSpan.Start++
				preEscape = true
			} else if line[0] == '#' {
				return
			} else if len(line) >= 2 && line[0] == '`' && line[1] == 't' {
				p.consumeTableFenceBlocks(doc, line, lineSpan, srcLine, s, collectDiag, diags)
				return
			} else if s.TableMode {
				s.TableLines = append(s.TableLines, line)
				return
			} else if len(line) >= 2 && line[0] == '`' && line[1] == '{' {
				pt := p.parsePartialFromInner(line[2:], s)
				if pt == nil {
					if collectDiag {
						*diags = append(*diags, Diagnostic{
							Severity: SeverityWarning,
							Code:     "partial.malformed",
							Message:  "malformed partial ignored",
							Span:     lineSpan,
						})
					}
					return
				}
				doc.Blocks = append(doc.Blocks, Block{
					Kind:       BlockPartial,
					Span:       lineSpan,
					SourceLine: srcLine,
					Depth:      s.Depth,
					Align:      s.Align,
					Partial:    pt,
				})
				return
			} else if line[0] == '<' {
				s.Depth = 0
				if len(line) == 1 {
					return
				}
				p.appendBlocksFromLine(doc, line[1:], Span{Start: lineSpan.Start + 1, End: lineSpan.End}, srcLine, s, collectDiag, diags)
				return
			} else if line[0] == '>' {
				i := 0
				for i < len(line) && line[i] == '>' {
					i++
				}
				if collectDiag && i > maxSectionDepth {
					*diags = append(*diags, Diagnostic{
						Severity: SeverityWarning,
						Code:     "section.depth_capped",
						Message:  "section depth capped at 16",
						Span:     Span{Start: lineSpan.Start, End: lineSpan.Start + i},
					})
				}
				s.Depth = capSectionDepth(i)
				headingLine := trimASCIISpaces(line[i:])
				if headingLine == "" {
					return
				}
				style := headingStyle(p, i)
				latched := p.stateToStyle(s)
				p.styleToState(style, s)
				parts := p.makeOutput(s, headingLine, false)
				p.styleToState(latched, s)
				inlines := partsToInlines(parts, lineSpan.Start+i)
				if !inlinesHaveContent(inlines) {
					p.styleToState(latched, s)
					return
				}
				doc.Blocks = append(doc.Blocks, Block{
					Kind:         BlockHeading,
					Span:         lineSpan,
					SourceLine:   srcLine,
					Depth:        s.Depth,
					HeadingStyle: style,
					Inlines:      inlines,
				})
				return
			} else if line[0] == '-' {
				if len(line) == 1 {
					doc.Blocks = append(doc.Blocks, Block{
						Kind:       BlockHR,
						Span:       lineSpan,
						SourceLine: srcLine,
						Depth:      s.Depth,
						LineBG:     s.BGColor,
						HeadingStyle: Style{
							FG: s.FGColor,
							BG: s.BGColor,
						},
					})
					return
				}
				_, firstSize := utf8.DecodeRuneInString(line)
				r, _ := utf8.DecodeRuneInString(line[firstSize:])
				doc.Blocks = append(doc.Blocks, Block{
					Kind:        BlockDivider,
					Span:        lineSpan,
					SourceLine:  srcLine,
					Depth:       s.Depth,
					LineBG:      s.BGColor,
					DividerRune: r,
					HeadingStyle: Style{
						FG: s.FGColor,
						BG: s.BGColor,
					},
				})
				return
			}
		}
		if !s.Literal && strings.IndexByte(line, '`') < 0 && !preEscape {
			parts := p.makeOutput(s, line, false)
			inlines := partsToInlines(parts, lineSpan.Start)
			if !inlinesHaveContent(inlines) {
				return
			}
			doc.Blocks = append(doc.Blocks, Block{
				Kind:       BlockParagraph,
				Span:       lineSpan,
				SourceLine: srcLine,
				Depth:      s.Depth,
				Align:      s.Align,
				LineBG:     s.BGColor,
				Inlines:    inlines,
			})
			return
		}
		if !p.ForceMonospace && s.Literal {
			text := line
			if line == "\\`=" {
				text = "`="
			}
			st := p.stateToStyle(s)
			doc.Blocks = append(doc.Blocks, Block{
				Kind:       BlockParagraph,
				Span:       lineSpan,
				SourceLine: srcLine,
				Depth:      s.Depth,
				Align:      s.Align,
				LineBG:     s.BGColor,
				Literal:    true,
				Inlines: []Inline{{
					Span:  lineSpan,
					Style: st,
					Text:  text,
				}},
			})
			return
		}
		parts := p.makeOutput(s, line, preEscape)
		inlines := partsToInlines(parts, lineSpan.Start)
		if !inlinesHaveContent(inlines) {
			return
		}
		b := Block{
			Kind:       BlockParagraph,
			Span:       lineSpan,
			SourceLine: srcLine,
			Depth:      s.Depth,
			Align:      s.Align,
			LineBG:     s.BGColor,
			Inlines:    inlines,
		}
		if s.Literal {
			b.Literal = true
		}
		doc.Blocks = append(doc.Blocks, b)
		return
	}
	doc.Blocks = append(doc.Blocks, Block{
		Kind:       BlockBlank,
		Span:       lineSpan,
		SourceLine: srcLine,
		Depth:      s.Depth,
		LineBG:     s.BGColor,
	})
}

func partsToInlines(parts []linePart, base int) []Inline {
	if len(parts) == 0 {
		return nil
	}
	out := make([]Inline, len(parts))
	off := base
	for i := range parts {
		pr := &parts[i]
		end := off
		if pr.text != "" {
			end = off + len(pr.text)
		}
		out[i] = Inline{
			Span:    Span{Start: off, End: end},
			Style:   pr.style,
			Text:    pr.text,
			Field:   pr.field,
			Link:    pr.link,
			Partial: pr.partial,
		}
		off = end
	}
	return out
}

func inlinesHaveContent(inlines []Inline) bool {
	for i := range inlines {
		in := &inlines[i]
		if in.Field != nil || in.Link != nil || in.Partial != nil {
			return true
		}
		if in.Text != "" {
			return true
		}
	}
	return false
}

func (p *Parser) consumeTableFenceBlocks(doc *Document, line string, lineSpan Span, srcLine int, s *State, collectDiag bool, diags *[]Diagnostic) {
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
			return
		}
		for _, ml := range micronLines {
			p.appendBlocksFromLine(doc, ml, lineSpan, srcLine, s, collectDiag, diags)
		}
		return
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
}

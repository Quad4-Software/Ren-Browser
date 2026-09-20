// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"io"
	"strings"
)

// ParseReader reads all bytes from r and parses them as Micron markup.
// Prefer Parse when the full string is already available.
func (p *Parser) ParseReader(r io.Reader) (*Document, error) {
	if r == nil {
		return p.Parse(""), nil
	}
	var b strings.Builder
	_, err := io.Copy(&b, r)
	if err != nil {
		return nil, err
	}
	return p.Parse(b.String()), nil
}

// ConvertReader reads markup from r and returns HTML.
func (p *Parser) ConvertReader(r io.Reader) (string, error) {
	doc, err := p.ParseReader(r)
	if err != nil {
		return "", err
	}
	return p.RenderHTML(doc), nil
}

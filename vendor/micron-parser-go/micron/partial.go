// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"strconv"
	"strings"
)

// parsePartialFromInner parses the interior after a line-leading "`{" up to the
// first '}', matching MicronParser.parse_partial in NomadNet and parsePartial in
// micron-parser-js (url[`refresh[`fields]]).
func (p *Parser) parsePartialFromInner(rest string, s *State) *Partial {
	before, _, ok := strings.Cut(rest, "}")
	if !ok {
		return nil
	}
	data := before
	if data == "" {
		return nil
	}
	parts := strings.Split(data, "`")
	if len(parts) > 3 {
		return nil
	}
	var urlPart string
	var fieldsStr string
	var refresh *float64
	switch len(parts) {
	case 1:
		urlPart = parts[0]
	case 2:
		urlPart = parts[0]
		if r, err := strconv.ParseFloat(parts[1], 64); err == nil {
			refresh = &r
		}
	case 3:
		urlPart = parts[0]
		if r, err := strconv.ParseFloat(parts[1], 64); err == nil {
			refresh = &r
		}
		fieldsStr = parts[2]
	}
	urlPart = strings.TrimSpace(urlPart)
	if urlPart == "" {
		return nil
	}
	pt := &Partial{
		URL:         FormatNomadnetworkURL(urlPart),
		Destination: urlPart,
		Descriptor:  data,
		Style:       p.stateToStyle(s),
	}
	if fieldsStr != "" {
		pt.FieldsAttr = fieldsStr
		for f := range strings.SplitSeq(fieldsStr, "|") {
			if strings.HasPrefix(f, "pid=") {
				pt.PartialID = f[len("pid="):]
				break
			}
		}
	}
	if refresh != nil && *refresh >= 1 {
		pt.HasRefresh = true
		pt.Refresh = *refresh
	}
	return pt
}

func formatPartialRefresh(r float64) string {
	return strconv.FormatFloat(r, 'f', -1, 64)
}

// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

// SemanticFingerprint is a JSON-serializable structural summary used for
// differential tests against the NomadNet-aligned Python oracle (primary dialect
// authority). Stock MicronParser emits Urwid widgets and is not invoked headlessly
// Parity is semantic (sections, links, fields, partials, headers), not widget trees.
type SemanticFingerprint struct {
	Authority string               `json:"authority"`
	Sections  []int                `json:"sections"`
	Links     []FingerprintLink    `json:"links"`
	Fields    []FingerprintField   `json:"fields"`
	Partials  []FingerprintPartial `json:"partials"`
	Headers   PageColors           `json:"headers"`
}

// FingerprintLink is a link entry in a semantic fingerprint.
type FingerprintLink struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}

// FingerprintField is a field entry in a semantic fingerprint.
type FingerprintField struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Data string `json:"data,omitempty"`
}

// FingerprintPartial is a partial entry in a semantic fingerprint.
type FingerprintPartial struct {
	Destination string `json:"destination"`
}

// Fingerprint builds a semantic fingerprint from a Document.
func (doc *Document) Fingerprint() SemanticFingerprint {
	fp := SemanticFingerprint{
		Authority: "go",
		Sections:  make([]int, 0),
		Links:     make([]FingerprintLink, 0),
		Fields:    make([]FingerprintField, 0),
		Partials:  make([]FingerprintPartial, 0),
		Headers:   doc.Colors,
	}
	if doc == nil {
		return fp
	}
	for i := range doc.Blocks {
		bl := &doc.Blocks[i]
		switch bl.Kind {
		case BlockHeading:
			fp.Sections = append(fp.Sections, bl.Depth)
		case BlockPartial:
			if bl.Partial != nil {
				fp.Partials = append(fp.Partials, FingerprintPartial{Destination: bl.Partial.Destination})
			}
		}
		collectInlineFingerprint(&fp, bl.Inlines)
	}
	return fp
}

func collectInlineFingerprint(fp *SemanticFingerprint, inlines []Inline) {
	for i := range inlines {
		in := &inlines[i]
		if in.Link != nil {
			fp.Links = append(fp.Links, FingerprintLink{URL: in.Link.URL, Label: stripHTMLRough(in.Link.Label)})
		}
		if in.Field != nil {
			kind := "text"
			switch in.Field.Kind {
			case FieldCheckbox:
				kind = "checkbox"
			case FieldRadio:
				kind = "radio"
			default:
				if in.Field.Masked {
					kind = "masked"
				}
			}
			data := in.Field.Value
			if in.Field.Kind == FieldCheckbox || in.Field.Kind == FieldRadio {
				data = in.Field.Label
			}
			fp.Fields = append(fp.Fields, FingerprintField{Name: in.Field.Name, Kind: kind, Data: data})
		}
		if in.Partial != nil {
			fp.Partials = append(fp.Partials, FingerprintPartial{Destination: in.Partial.Destination})
		}
	}
}

func stripHTMLRough(s string) string {
	if s == "" || s[0] != '<' {
		return s
	}
	var b []byte
	inTag := false
	for i := range len(s) {
		c := s[i]
		if c == '<' {
			inTag = true
			continue
		}
		if c == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b = append(b, c)
		}
	}
	return string(b)
}

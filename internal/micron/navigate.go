package micron

import (
	"strings"

	mp "micron-parser-go/micron"

	"renbrowser/internal/nomadnet"
)

type FieldInput struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Checked bool   `json:"checked"`
}

func HeaderColors(source string) (fg, bg string) {
	pc := mp.ParseHeaderTags(source)
	return pc.FG, pc.BG
}

func ResolveNavigation(currentURL, destination, fieldsSpec string, inputs []FieldInput) (string, error) {
	mpInputs := make([]mp.FieldInput, len(inputs))
	for i, in := range inputs {
		mpInputs[i] = mp.FieldInput{
			Type:    in.Type,
			Name:    in.Name,
			Value:   in.Value,
			Checked: in.Checked,
		}
	}
	allFields := mp.CollectFormFields(mpInputs)
	payload := mp.BuildRequestPayload(allFields, destination, fieldsSpec)

	req := nomadnet.RequestData{
		Vars:   payload.RequestVars,
		Fields: payload.Fields,
	}

	dest := strings.TrimSpace(payload.Destination)
	nodeHash := ""
	path := dest

	if len(dest) >= 33 && dest[32] == ':' && isHex32(dest[:32]) {
		nodeHash = strings.ToLower(dest[:32])
		path = dest[33:]
	} else if strings.HasPrefix(dest, ":") {
		path = strings.TrimPrefix(dest, ":")
	}

	if nodeHash == "" {
		parsed, err := nomadnet.ParseURL(currentURL)
		if err != nil {
			return "", err
		}
		nodeHash = parsed.NodeHash
	}

	return nomadnet.FormatURLWithRequest(nodeHash, path, req), nil
}

func isHex32(s string) bool {
	if len(s) != 32 {
		return false
	}
	for i := 0; i < 32; i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

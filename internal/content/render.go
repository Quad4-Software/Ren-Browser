// SPDX-License-Identifier: MIT
package content

import (
	"html"
	"strings"

	"renbrowser/internal/nomadnet"
	"renbrowser/internal/plugins"
)

var defaultRegistry = plugins.NewRegistry()

func SetRendererRegistry(reg *plugins.Registry) {
	if reg != nil {
		defaultRegistry = reg
	}
}

func RendererRegistry() *plugins.Registry {
	return defaultRegistry
}

func Render(path string, body []byte, nodeHash string) Rendered {
	raw := string(body)
	kind := nomadnet.DetectContentType(path, body)

	if out, ok := defaultRegistry.Render(path, body, nodeHash, kind); ok {
		return Rendered{
			Kind:    firstNonEmpty(out.Kind, kind),
			HTML:    out.HTML,
			Raw:     firstNonEmpty(out.Raw, raw),
			PageFG:  out.PageFG,
			PageBG:  out.PageBG,
			IsError: out.IsError,
		}
	}

	return Rendered{
		Kind: string(nomadnet.KindPlaintext),
		HTML: `<pre class="plaintext">` + html.EscapeString(raw) + `</pre>`,
		Raw:  raw,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

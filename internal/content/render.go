// SPDX-License-Identifier: MIT
package content

import (
	"html"
	"strings"

	"renbrowser/internal/micron"
	"renbrowser/internal/nomadnet"
)

type Rendered struct {
	Kind    string `json:"kind"`
	HTML    string `json:"html"`
	Raw     string `json:"raw"`
	PageFG  string `json:"pageFg,omitempty"`
	PageBG  string `json:"pageBg,omitempty"`
	IsError bool   `json:"isError"`
}

func Render(path string, body []byte, nodeHash string) Rendered {
	raw := string(body)
	kind := nomadnet.DetectContentType(path, body)

	switch nomadnet.ContentKind(kind) {
	case nomadnet.KindMicron:
		htmlOut, fg, bg := micron.RenderDark(raw)
		if nodeHash != "" {
			htmlOut = isolateNomadLinks(htmlOut, nodeHash)
		}
		return Rendered{Kind: kind, HTML: htmlOut, Raw: raw, PageFG: fg, PageBG: bg}
	case nomadnet.KindHTML:
		return Rendered{Kind: kind, HTML: SanitizeHTML(raw), Raw: raw}
	case nomadnet.KindMarkdown:
		return Rendered{Kind: kind, HTML: markdownToHTML(raw), Raw: raw}
	case nomadnet.KindPlaintext:
		return Rendered{
			Kind: kind,
			HTML: `<pre class="plaintext">` + html.EscapeString(raw) + `</pre>`,
			Raw:  raw,
		}
	default:
		return Rendered{
			Kind: string(nomadnet.KindPlaintext),
			HTML: `<pre class="plaintext">` + html.EscapeString(raw) + `</pre>`,
			Raw:  raw,
		}
	}
}

func markdownToHTML(md string) string {
	var b strings.Builder
	b.Grow(len(md) + len(md)/4 + 64)
	b.WriteString(`<article class="markdown">`)
	inCode := false

	for line := range strings.SplitSeq(md, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			if inCode {
				b.WriteString(`</code></pre>`)
				inCode = false
			} else {
				b.WriteString(`<pre class="md-pre"><code>`)
				inCode = true
			}
			continue
		}
		if inCode {
			b.WriteString(html.EscapeString(line))
			b.WriteByte('\n')
			continue
		}
		switch {
		case strings.HasPrefix(trim, "### "):
			writeMDBlock(&b, "h3", strings.TrimPrefix(trim, "### "))
		case strings.HasPrefix(trim, "## "):
			writeMDBlock(&b, "h2", strings.TrimPrefix(trim, "## "))
		case strings.HasPrefix(trim, "# "):
			writeMDBlock(&b, "h1", strings.TrimPrefix(trim, "# "))
		case strings.HasPrefix(trim, "- "):
			b.WriteString(`<li>`)
			b.WriteString(html.EscapeString(strings.TrimPrefix(trim, "- ")))
			b.WriteString(`</li>`)
		case trim == "":
			continue
		default:
			b.WriteString(`<p>`)
			b.WriteString(html.EscapeString(trim))
			b.WriteString(`</p>`)
		}
	}
	if inCode {
		b.WriteString(`</code></pre>`)
	}
	b.WriteString(`</article>`)
	return b.String()
}

func writeMDBlock(b *strings.Builder, tag, text string) {
	b.WriteString("<")
	b.WriteString(tag)
	b.WriteString(">")
	b.WriteString(html.EscapeString(text))
	b.WriteString("</")
	b.WriteString(tag)
	b.WriteString(">")
}

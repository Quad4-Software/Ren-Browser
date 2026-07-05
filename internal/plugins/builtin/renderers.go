// SPDX-License-Identifier: MIT
package builtin

import (
	"html"
	"strings"

	"renbrowser/internal/content"
	"renbrowser/internal/micron"
	"renbrowser/internal/nomadnet"
	"renbrowser/internal/plugins"
)

type micronRenderer struct{}

func (micronRenderer) ID() string       { return "builtin.micron" }
func (micronRenderer) Priority() int    { return 100 }
func (micronRenderer) PluginID() string { return "" }
func (micronRenderer) Match(path string, body []byte, detected string) bool {
	return nomadnet.ContentKind(detected) == nomadnet.KindMicron
}
func (r micronRenderer) Render(path string, body []byte, nodeHash string) (plugins.Rendered, error) {
	raw := string(body)
	htmlOut, fg, bg := micron.RenderDark(raw)
	if nodeHash != "" {
		htmlOut = content.IsolateNomadLinks(htmlOut, nodeHash)
	}
	return plugins.Rendered{Kind: string(nomadnet.KindMicron), HTML: htmlOut, Raw: raw, PageFG: fg, PageBG: bg}, nil
}

type htmlRenderer struct{}

func (htmlRenderer) ID() string       { return "builtin.html" }
func (htmlRenderer) Priority() int    { return 90 }
func (htmlRenderer) PluginID() string { return "" }
func (htmlRenderer) Match(path string, body []byte, detected string) bool {
	return nomadnet.ContentKind(detected) == nomadnet.KindHTML
}
func (htmlRenderer) Render(path string, body []byte, nodeHash string) (plugins.Rendered, error) {
	raw := string(body)
	return plugins.Rendered{Kind: string(nomadnet.KindHTML), HTML: content.SanitizeHTML(raw), Raw: raw}, nil
}

type markdownRenderer struct{}

func (markdownRenderer) ID() string       { return "builtin.markdown" }
func (markdownRenderer) Priority() int    { return 80 }
func (markdownRenderer) PluginID() string { return "" }
func (markdownRenderer) Match(path string, body []byte, detected string) bool {
	return nomadnet.ContentKind(detected) == nomadnet.KindMarkdown
}
func (markdownRenderer) Render(path string, body []byte, nodeHash string) (plugins.Rendered, error) {
	raw := string(body)
	return plugins.Rendered{Kind: string(nomadnet.KindMarkdown), HTML: markdownToHTML(raw), Raw: raw}, nil
}

type plaintextRenderer struct{}

func (plaintextRenderer) ID() string       { return "builtin.plaintext" }
func (plaintextRenderer) Priority() int    { return 10 }
func (plaintextRenderer) PluginID() string { return "" }
func (plaintextRenderer) Match(path string, body []byte, detected string) bool {
	kind := nomadnet.ContentKind(detected)
	return kind == nomadnet.KindPlaintext || kind == ""
}
func (plaintextRenderer) Render(path string, body []byte, nodeHash string) (plugins.Rendered, error) {
	raw := string(body)
	return plugins.Rendered{
		Kind: string(nomadnet.KindPlaintext),
		HTML: `<pre class="plaintext">` + html.EscapeString(raw) + `</pre>`,
		Raw:  raw,
	}, nil
}

func RegisterRenderers(reg *plugins.Registry) {
	reg.RegisterRenderer(micronRenderer{})
	reg.RegisterRenderer(htmlRenderer{})
	reg.RegisterRenderer(markdownRenderer{})
	reg.RegisterRenderer(plaintextRenderer{})
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

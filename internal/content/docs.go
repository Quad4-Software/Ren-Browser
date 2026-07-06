// SPDX-License-Identifier: MIT

package content

import (
	"fmt"
	"html"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strings"

	docassets "renbrowser/docs"
)

var (
	supportedDocLangs = []string{"en", "ru", "es", "de"}
	docsPageNameRe    = regexp.MustCompile(`^[a-z0-9-]+$`)
)

type DocsRenderInput struct {
	RawURL    string
	SavedLang string
	SaveLang  func(lang string)
}

type DocsRenderResult struct {
	URL          string
	HTML         string
	Raw          string
	HistoryTitle string
}

func RenderDocs(in DocsRenderInput) (DocsRenderResult, bool) {
	if !MatchDocsURL(in.RawURL) {
		return DocsRenderResult{}, false
	}
	lang, page := ParseDocsQuery(in.RawURL)
	if lang != "" && ValidDocsLang(lang) {
		if in.SaveLang != nil {
			in.SaveLang(lang)
		}
	} else if in.SavedLang != "" && ValidDocsLang(in.SavedLang) {
		lang = in.SavedLang
	} else {
		return DocsRenderResult{
			URL:          "docs:",
			HTML:         renderDocsLanguagePicker(),
			HistoryTitle: "Documentation",
		}, true
	}

	body, title, err := loadDocsMarkdown(lang, page)
	if err != nil {
		return DocsRenderResult{
			URL:          FormatDocsURL(lang, page),
			HTML:         renderDocsError(err),
			HistoryTitle: "Documentation",
		}, true
	}

	return DocsRenderResult{
		URL:          FormatDocsURL(lang, page),
		Raw:          body,
		HistoryTitle: title,
	}, true
}

func SupportedDocsLangs() []string {
	out := make([]string, len(supportedDocLangs))
	copy(out, supportedDocLangs)
	return out
}

func ValidDocsLang(lang string) bool {
	lang = strings.ToLower(strings.TrimSpace(lang))
	return slices.Contains(supportedDocLangs, lang)
}

func MatchDocsURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case lower == "docs", lower == "docs:":
		return true
	case strings.HasPrefix(lower, "docs?"):
		return true
	case strings.HasPrefix(lower, "docs:?"):
		return true
	default:
		return false
	}
}

func ParseDocsQuery(raw string) (lang, page string) {
	raw = strings.TrimSpace(raw)
	_, after, ok := strings.Cut(raw, "?")
	if !ok {
		return "", ""
	}
	values, err := url.ParseQuery(after)
	if err != nil {
		return "", ""
	}
	return strings.ToLower(strings.TrimSpace(values.Get("lang"))), SanitizeDocsPage(values.Get("page"))
}

func FormatDocsURL(lang, page string) string {
	if lang == "" {
		return "docs:"
	}
	values := url.Values{}
	values.Set("lang", lang)
	if page != "" {
		values.Set("page", page)
	}
	return "docs:?" + values.Encode()
}

func SanitizeDocsPage(page string) string {
	page = strings.TrimSpace(page)
	page = strings.TrimSuffix(page, ".md")
	page = strings.ToLower(page)
	if page == "" || page == "readme" {
		return ""
	}
	if !docsPageNameRe.MatchString(page) {
		return ""
	}
	return page
}

func loadDocsMarkdown(lang, page string) (body string, title string, err error) {
	file := "README.md"
	if page != "" {
		file = page + ".md"
	}
	data, err := docassets.FS.ReadFile(path.Join(lang, file))
	if err != nil {
		return "", "", fmt.Errorf("documentation page not found")
	}
	body = string(data)
	title = docsTitleFromMarkdown(body, page)
	return body, title, nil
}

func docsTitleFromMarkdown(body, page string) string {
	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(after)
		}
	}
	if page != "" {
		return strings.ReplaceAll(page, "-", " ")
	}
	return "Documentation"
}

func renderDocsLanguagePicker() string {
	var links strings.Builder
	for _, lang := range supportedDocLangs {
		label := docsLangLabel(lang)
		links.WriteString(`<li><a href="`)
		links.WriteString(html.EscapeString(FormatDocsURL(lang, "")))
		links.WriteString(`">`)
		links.WriteString(html.EscapeString(label))
		links.WriteString(`</a></li>`)
	}
	return `<article class="docs-page docs-picker">` +
		`<h1>Documentation</h1>` +
		`<p>Choose a language. Your choice is remembered for next time.</p>` +
		`<ul class="docs-lang-list">` + links.String() + `</ul>` +
		`<p class="docs-hint">You can also open <code>docs:?lang=en</code> or <code>docs:?lang=ru&amp;page=faq</code> in the address bar.</p>` +
		`</article>`
}

func docsLangLabel(lang string) string {
	switch lang {
	case "en":
		return "English"
	case "ru":
		return "Russian"
	case "es":
		return "Spanish"
	case "de":
		return "German"
	default:
		return strings.ToUpper(lang)
	}
}

func renderDocsError(err error) string {
	return `<article class="docs-page docs-error"><h1>Documentation</h1><p>` +
		html.EscapeString(err.Error()) +
		`</p><p><a href="docs:">Back to language list</a></p></article>`
}

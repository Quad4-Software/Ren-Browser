// SPDX-License-Identifier: MIT
package content

import (
	"regexp"
	"strings"
)

var (
	hexHashRe  = regexp.MustCompile(`^[a-f0-9]{32}$`)
	anchorHref = regexp.MustCompile(`<a\s+([^>]*?)href="([^"]*)"([^>]*)>`)
)

func IsolateNomadLinks(html, nodeHash string) string {
	return isolateNomadLinks(html, nodeHash)
}

func isolateNomadLinks(html, nodeHash string) string {
	if html == "" || nodeHash == "" {
		return html
	}
	nodeHash = strings.ToLower(nodeHash)
	matches := anchorHref.FindAllStringSubmatchIndex(html, -1)
	if len(matches) == 0 {
		return html
	}

	var b strings.Builder
	b.Grow(len(html) + len(matches)*24)
	last := 0
	for _, m := range matches {
		if len(m) < 8 {
			continue
		}
		b.WriteString(html[last:m[0]])
		href := strings.TrimSpace(html[m[4]:m[5]])
		if href == "" || href == "#" || strings.HasPrefix(href, "#") {
			b.WriteString(html[m[0]:m[1]])
			last = m[1]
			continue
		}
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			b.WriteString(html[m[0]:m[1]])
			last = m[1]
			continue
		}
		full := resolveNomadHref(nodeHash, href)
		if full == "" {
			b.WriteString(`<a href="#" `)
			b.WriteString(html[m[2]:m[3]])
			b.WriteString(html[m[6]:m[7]])
			b.WriteByte('>')
		} else {
			b.WriteString(`<a href="#" data-nomad-url="`)
			b.WriteString(full)
			b.WriteString(`" `)
			b.WriteString(html[m[2]:m[3]])
			b.WriteString(html[m[6]:m[7]])
			b.WriteByte('>')
		}
		last = m[1]
	}
	b.WriteString(html[last:])
	return b.String()
}

func resolveNomadHref(nodeHash, href string) string {
	href = strings.TrimSpace(href)
	if len(href) >= 33 && href[32] == ':' && hexHashRe.MatchString(strings.ToLower(href[:32])) {
		return strings.ToLower(href[:32]) + ":" + normalizeNomadPath(href[33:])
	}
	if strings.HasPrefix(href, "/page/") || strings.HasPrefix(href, "/file/") {
		return nodeHash + ":" + href
	}
	if strings.HasPrefix(href, ":") {
		return nodeHash + ":" + normalizeNomadPath(strings.TrimPrefix(href, ":"))
	}
	return ""
}

func normalizeNomadPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/page/index.mu"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

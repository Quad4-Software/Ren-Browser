// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package micron

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Micron image limits mirror MeshChatX MicronParser.js.
const (
	micronImageMaxWidth     = 8192
	micronImageMaxHeight    = 8192
	micronImageMaxSizeHint  = 100 * 1024 * 1024 // 100 MiB
	micronImageMaxAltLen    = 240
	micronImageMaxKeyLen    = 64
	micronImageMaxProfileLn = 32
)

var (
	micronImageKeyRe     = regexp.MustCompile(`^[a-zA-Z0-9_.-]*$`)
	micronImageProfileRe = regexp.MustCompile(`^[a-zA-Z0-9_-]*$`)
	micronImageHashRe    = regexp.MustCompile(`(?i)^[a-f0-9]{32}$`)
	micronImageMediaRe   = regexp.MustCompile(`(?i)\.(webp|png|jpe?g|bmp|gif|tiff)$`)
)

// imageOptions carries the parsed key=value image hints from link fields.
type imageOptions struct {
	img     bool
	w       int
	h       int
	size    int
	key     string
	align   string
	profile string
}

// LinkImage describes a link that renders as a deferred image placeholder
// instead of an anchor. It is set on Link.Image only when detection succeeds.
type LinkImage struct {
	RawURL  string `json:"raw_url"`
	Path    string `json:"path"`
	Alt     string `json:"alt"`
	Width   int    `json:"w,omitempty"`
	Height  int    `json:"h,omitempty"`
	Size    int    `json:"size,omitempty"`
	Key     string `json:"key,omitempty"`
	Align   string `json:"align,omitempty"`
	Profile string `json:"profile,omitempty"`
}

// clampMicronImageNumber parses a positive finite number, floors it, and
// clamps it to limit. It returns ok=false for anything else, matching the JS
// Number() based parsing in parseMicronImageOptions.
func clampMicronImageNumber(v string, limit int) (int, bool) {
	n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || n <= 0 {
		return 0, false
	}
	f := math.Floor(n)
	if f >= float64(limit) {
		return limit, true
	}
	return int(f), true
}

// sanitizeMicronImageString trims, truncates to maxLen runes, then validates
// against re. An empty result means the value is rejected.
func sanitizeMicronImageString(v string, maxLen int, re *regexp.Regexp) string {
	trimmed := strings.TrimSpace(v)
	if utf8.RuneCountInString(trimmed) > maxLen {
		trimmed = string([]rune(trimmed)[:maxLen])
	}
	if re != nil && !re.MatchString(trimmed) {
		return ""
	}
	return trimmed
}

// truncateMicronImageAlt trims whitespace and caps the alt text at
// micronImageMaxAltLen runes.
func truncateMicronImageAlt(v string) string {
	trimmed := strings.TrimSpace(v)
	if utf8.RuneCountInString(trimmed) > micronImageMaxAltLen {
		trimmed = string([]rune(trimmed)[:micronImageMaxAltLen])
	}
	return trimmed
}

// parseMicronImageOptions scans pipe-split link fields. Each field may carry
// several semicolon separated key=value parts, like the JS implementation.
func parseMicronImageOptions(fields []string) imageOptions {
	opts := imageOptions{align: "left"}
	for _, raw := range fields {
		if raw == "" {
			continue
		}
		for part := range strings.SplitSeq(raw, ";") {
			idx := strings.IndexByte(part, '=')
			if idx <= 0 {
				continue
			}
			k := strings.ToLower(strings.TrimSpace(part[:idx]))
			v := strings.TrimSpace(part[idx+1:])
			switch k {
			case "img":
				switch strings.ToLower(v) {
				case "1", "true", "yes":
					opts.img = true
				default:
					opts.img = false
				}
			case "w":
				if n, ok := clampMicronImageNumber(v, micronImageMaxWidth); ok {
					opts.w = n
				} else {
					opts.w = 0
				}
			case "h":
				if n, ok := clampMicronImageNumber(v, micronImageMaxHeight); ok {
					opts.h = n
				} else {
					opts.h = 0
				}
			case "s":
				if n, ok := clampMicronImageNumber(v, micronImageMaxSizeHint); ok {
					opts.size = n
				} else {
					opts.size = 0
				}
			case "k":
				opts.key = sanitizeMicronImageString(v, micronImageMaxKeyLen, micronImageKeyRe)
			case "a":
				switch strings.ToLower(v) {
				case "left", "l":
					opts.align = "left"
				case "center", "c":
					opts.align = "center"
				case "right", "r":
					opts.align = "right"
				}
			case "profile":
				opts.profile = sanitizeMicronImageString(v, micronImageMaxProfileLn, micronImageProfileRe)
			}
		}
	}
	return opts
}

// extractMicronImageFilePath normalizes a node file URL into hash:/path or
// :/path form. It returns an empty string when the URL is not a safe media or
// file path, matching MeshChatX MicronParser.extractMicronImageFilePath.
func extractMicronImageFilePath(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	url := rawURL
	if len(url) >= len("nomadnetwork://") && strings.EqualFold(url[:len("nomadnetwork://")], "nomadnetwork://") {
		url = url[len("nomadnetwork://"):]
	}
	if i := strings.IndexByte(url, '`'); i >= 0 {
		url = url[:i]
	}
	if i := strings.IndexByte(url, '?'); i >= 0 {
		url = url[:i]
	}
	if i := strings.IndexByte(url, '#'); i >= 0 {
		url = url[:i]
	}
	url = strings.TrimSpace(url)

	path := url
	var hash string
	if i := strings.Index(url, ":/"); i >= 0 {
		hash = url[:i]
		path = url[i+2:]
		if hash != "" && !micronImageHashRe.MatchString(hash) {
			return ""
		}
	} else if strings.HasPrefix(url, ":") {
		path = url[1:]
	}

	if strings.HasPrefix(path, "media/") {
		if !micronImageMediaRe.MatchString(path) {
			return ""
		}
	} else if strings.HasPrefix(path, "file/") {
		if !strings.HasSuffix(strings.ToLower(path), ".webp") {
			return ""
		}
	} else {
		return ""
	}

	if strings.Contains(path, "..") {
		return ""
	}
	for _, r := range path {
		if r < 32 || strings.ContainsRune(`<>"|?*`, r) {
			return ""
		}
	}

	if hash != "" {
		return hash + ":/" + path
	}
	return ":/" + path
}

// detectImage mirrors the MeshChatX parseLink override: with fields present, a
// link becomes an image when the img option is set or the normalized path is
// under media/. A non-empty alt text and a valid image path are required.
func (lk *Link) detectImage(rawLabel string) {
	if len(lk.Fields) == 0 {
		return
	}
	opts := parseMicronImageOptions(lk.Fields)
	rawURL := strings.TrimPrefix(lk.URL, "nomadnetwork://")
	imagePath := extractMicronImageFilePath(rawURL)
	withoutHash := imagePath
	if i := strings.Index(withoutHash, ":/"); i >= 0 {
		withoutHash = withoutHash[i+2:]
	}
	isMedia := strings.HasPrefix(withoutHash, "/media/") || strings.HasPrefix(withoutHash, "media/")
	if !opts.img && !isMedia {
		return
	}
	alt := truncateMicronImageAlt(rawLabel)
	if imagePath == "" || alt == "" {
		return
	}
	lk.Image = &LinkImage{
		RawURL:  rawURL,
		Path:    imagePath,
		Alt:     alt,
		Width:   opts.w,
		Height:  opts.h,
		Size:    opts.size,
		Key:     opts.key,
		Align:   opts.align,
		Profile: opts.profile,
	}
}

// formatMicronImageSize renders a byte count like the JS Intl based helper:
// raw bytes under 1 KiB, then one-decimal kB or MB with the fraction dropped
// when it rounds to zero.
func formatMicronImageSize(bytes int) string {
	if bytes < 0 {
		return ""
	}
	if bytes < 1024 {
		return strconv.Itoa(bytes) + " B"
	}
	if bytes < 1024*1024 {
		return formatMicronImageDecimal(float64(bytes)/1024) + " kB"
	}
	return formatMicronImageDecimal(float64(bytes)/(1024*1024)) + " MB"
}

func formatMicronImageDecimal(v float64) string {
	r := math.Round(v*10) / 10
	return strconv.FormatFloat(r, 'f', -1, 64)
}

// SPDX-License-Identifier: MIT
package plugins

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func CollectPluginI18nLocales(dir string, embedded map[string][]byte) []string {
	seen := make(map[string]struct{})
	addFile := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		base := filepath.Base(strings.ReplaceAll(name, `\`, "/"))
		if !strings.HasSuffix(strings.ToLower(base), ".json") {
			return
		}
		code := strings.TrimSuffix(base, filepath.Ext(base))
		code = strings.TrimSpace(code)
		if code == "" {
			return
		}
		seen[code] = struct{}{}
	}

	for path := range embedded {
		clean := filepath.ToSlash(filepath.Clean(strings.ReplaceAll(path, `\`, "/")))
		if strings.HasPrefix(clean, "locales/") {
			addFile(clean)
		}
	}

	if dir != "" {
		localesDir := filepath.Join(dir, "locales")
		entries, err := os.ReadDir(localesDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				addFile(entry.Name())
			}
		}
	}

	out := make([]string, 0, len(seen))
	for code := range seen {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

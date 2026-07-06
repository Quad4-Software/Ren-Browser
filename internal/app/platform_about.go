// SPDX-License-Identifier: MIT
package app

import (
	"fmt"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"

	"renbrowser/internal/content"
)

func platformRuntimeRows(app *application.App) []content.AboutRow {
	if app == nil || app.Env == nil {
		return nil
	}
	env := app.Env.Info()
	var rows []content.AboutRow

	pi := env.PlatformInfo
	if pi == nil {
		return appendWebEngineFallback(rows, env.OS)
	}

	if mode := platformString(pi["mode"]); mode == "server" {
		return appendRow(rows, "Mode", "Headless server")
	}

	switch env.OS {
	case "linux":
		gtkRuntime := firstPlatformString(pi, "gtk4-runtime", "gtk3-runtime")
		webkitRuntime := firstPlatformString(pi, "webkitgtk6-runtime", "webkit2gtk-runtime")
		rows = appendRow(rows, "GTK", gtkRuntime)
		rows = appendRow(rows, "WebKitGTK", webkitRuntime)
		if gtkCompiled := firstPlatformString(pi, "gtk4-compiled", "gtk3-compiled"); gtkCompiled != "" && gtkCompiled != gtkRuntime {
			rows = appendRow(rows, "GTK (built with)", gtkCompiled)
		}
		if webkitCompiled := firstPlatformString(pi, "webkitgtk6-compiled", "webkit2gtk-compiled"); webkitCompiled != "" && webkitCompiled != webkitRuntime {
			rows = appendRow(rows, "WebKitGTK (built with)", webkitCompiled)
		}
		if wayland, ok := pi["wayland"].(bool); ok {
			if wayland {
				rows = appendRow(rows, "Display server", "Wayland")
			} else {
				rows = appendRow(rows, "Display server", "X11")
			}
		}
		rows = appendRow(rows, "Compositor", platformString(pi["compositor"]))
	case "windows":
		rows = appendRow(rows, "WebView2", platformString(pi["WebView2"]))
		if loader, ok := pi["Go-WebView2Loader"].(bool); ok {
			rows = appendRow(rows, "WebView2 loader", map[bool]string{true: "Go loader", false: "Native loader"}[loader])
		}
	case "darwin":
		rows = appendRow(rows, "Web engine", "System WebKit")
	case "android":
		rows = appendRow(rows, "Platform", platformString(pi["platform"]))
		rows = appendRow(rows, "Web engine", "Android System WebView")
	default:
		rows = appendWebEngineFallback(rows, env.OS)
	}

	return rows
}

func appendWebEngineFallback(rows []content.AboutRow, os string) []content.AboutRow {
	if len(rows) > 0 || strings.TrimSpace(os) == "" {
		return rows
	}
	return appendRow(rows, "Web engine", fmt.Sprintf("Browser shell (%s)", os))
}

func firstPlatformString(pi map[string]any, keys ...string) string {
	for _, key := range keys {
		if v := platformString(pi[key]); v != "" {
			return v
		}
	}
	return ""
}

func platformString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case bool:
		if t {
			return "yes"
		}
		return "no"
	default:
		return ""
	}
}

func appendRow(rows []content.AboutRow, label, value string) []content.AboutRow {
	value = strings.TrimSpace(value)
	if value == "" {
		return rows
	}
	return append(rows, content.AboutRow{Label: label, Value: value})
}

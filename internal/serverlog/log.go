//go:build server

// SPDX-License-Identifier: MIT

package serverlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

const (
	colorReset  = "\033[0m"
	colorDim    = "\033[2m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
)

var (
	mu      sync.Mutex
	useANSI bool
	out     io.Writer = os.Stderr
)

func Init() {
	useANSI = ansiEnabled()
}

func Slog() *slog.Logger {
	return slog.New(&handler{w: out, color: useANSI})
}

func Banner(title, version, commit string) {
	mu.Lock()
	defer mu.Unlock()

	line := strings.Repeat("─", 44)
	if useANSI {
		fmt.Fprintf(out, "\n%s%s%s %s%s%s\n", colorBold, colorCyan, title, version, colorDim, " ("+commit+")"+colorReset+"\n")
		fmt.Fprintf(out, "%s%s%s\n\n", colorDim, line, colorReset)
		return
	}
	fmt.Fprintf(out, "\n%s %s (%s)\n%s\n\n", title, version, commit, line)
}

func Field(key, value string) {
	mu.Lock()
	defer mu.Unlock()

	if value == "" {
		value = "—"
	}
	if useANSI {
		fmt.Fprintf(out, "  %s%-10s%s %s\n", colorDim, key, colorReset, value)
		return
	}
	fmt.Fprintf(out, "  %-10s %s\n", key+":", value)
}

func OK(message string) {
	mu.Lock()
	defer mu.Unlock()

	if useANSI {
		fmt.Fprintf(out, "%s ok %s%s %s\n", colorGreen, colorReset, colorBold, message+colorReset)
		return
	}
	fmt.Fprintf(out, "[ok] %s\n", message)
}

func Info(message string) {
	mu.Lock()
	defer mu.Unlock()

	if useANSI {
		fmt.Fprintf(out, "%s..%s %s\n", colorBlue, colorReset, message)
		return
	}
	fmt.Fprintf(out, "[info] %s\n", message)
}

func Warn(message string) {
	mu.Lock()
	defer mu.Unlock()

	if useANSI {
		fmt.Fprintf(out, "%s!!%s %s\n", colorYellow, colorReset, message)
		return
	}
	fmt.Fprintf(out, "[warn] %s\n", message)
}

func Error(message string, err error) {
	mu.Lock()
	defer mu.Unlock()

	detail := ""
	if err != nil {
		detail = err.Error()
	}
	if useANSI {
		if detail != "" {
			fmt.Fprintf(out, "%s!!%s %s: %s%s\n", colorRed, colorReset, message, colorDim, detail+colorReset)
			return
		}
		fmt.Fprintf(out, "%s!!%s %s\n", colorRed, colorReset, message)
		return
	}
	if detail != "" {
		fmt.Fprintf(out, "[error] %s: %s\n", message, detail)
		return
	}
	fmt.Fprintf(out, "[error] %s\n", message)
}

func Ready(listenURL string) {
	mu.Lock()
	defer mu.Unlock()

	if useANSI {
		fmt.Fprintf(out, "\n%s ready %s%s %s\n\n", colorGreen, colorReset, colorBold, listenURL+colorReset)
		return
	}
	fmt.Fprintf(out, "\nready %s\n\n", listenURL)
}

func Emit(level, message, detail string) {
	switch level {
	case "error":
		if detail != "" {
			Error(message, fmt.Errorf("%s", detail))
		} else {
			Error(message, nil)
		}
	case "warn", "warning":
		if detail != "" {
			Warn(message + ": " + detail)
		} else {
			Warn(message)
		}
	case "info":
		if skipInfo(message) {
			return
		}
		if detail != "" {
			Info(message + ": " + detail)
		} else {
			Info(message)
		}
	}
}

func skipInfo(message string) bool {
	if os.Getenv("REN_BROWSER_VERBOSE") == "1" {
		return false
	}
	switch message {
	case "page loaded", "page cache hit", "log level":
		return true
	default:
		return false
	}
}

func ansiEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

type handler struct {
	w     io.Writer
	color bool
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	if level == slog.LevelDebug && os.Getenv("REN_BROWSER_VERBOSE") != "1" {
		return false
	}
	return true
}

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	mu.Lock()
	defer mu.Unlock()

	if h.color {
		level := colorDim + r.Level.String() + colorReset
		switch r.Level {
		case slog.LevelError:
			level = colorRed + "error" + colorReset
		case slog.LevelWarn:
			level = colorYellow + "warn" + colorReset
		case slog.LevelInfo:
			level = colorBlue + "info" + colorReset
		}
		msg := r.Message
		r.Attrs(func(a slog.Attr) bool {
			msg += fmt.Sprintf(" %s=%v", a.Key, a.Value)
			return true
		})
		fmt.Fprintf(h.w, "%s %s\n", level, msg)
		return nil
	}

	msg := r.Level.String() + " " + r.Message
	r.Attrs(func(a slog.Attr) bool {
		msg += fmt.Sprintf(" %s=%v", a.Key, a.Value)
		return true
	})
	fmt.Fprintf(h.w, "%s\n", msg)
	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	_ = attrs
	return h
}

func (h *handler) WithGroup(name string) slog.Handler {
	_ = name
	return h
}

//go:build !server

// SPDX-License-Identifier: MIT

package serverlog

import "log/slog"

func Init() {}

func Slog() *slog.Logger { return nil }

func Banner(string, string, string) {}

func Field(string, string) {}

func OK(string) {}

func Info(string) {}

func Warn(string) {}

func Error(string, error) {}

func Ready(string) {}

func Emit(string, string, string) {}

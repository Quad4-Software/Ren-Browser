//go:build !server

// SPDX-License-Identifier: MIT

package serverlog

import "log/slog"

func WailsLogger() *slog.Logger {
	return nil
}

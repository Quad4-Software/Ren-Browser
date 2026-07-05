package apperrors

import (
	"errors"
	"io/fs"
	"strings"
	"syscall"
)

type Kind string

const (
	KindConnectionFailed Kind = "connection_failed"
	KindConnectionLost   Kind = "connection_lost"
	KindNotFound         Kind = "not_found"
	KindInternal         Kind = "internal"
	KindStorageFull      Kind = "storage_full"
	KindDatabaseCorrupt  Kind = "database_corrupt"
	KindUnknown          Kind = "unknown"
)

func ClassifyFetch(errMsg string, body []byte) (Kind, string) {
	msg := strings.ToLower(strings.TrimSpace(errMsg))
	if msg == "" && len(body) > 0 {
		return classifyBody(body)
	}
	if msg == "" {
		return KindUnknown, ""
	}

	switch {
	case msg == "reticulum not ready":
		return KindConnectionFailed, errMsg
	case strings.Contains(msg, "empty response"):
		return KindNotFound, errMsg
	case strings.Contains(msg, "not found"), strings.Contains(msg, "404"):
		return KindNotFound, errMsg
	case strings.Contains(msg, "no path"),
		strings.Contains(msg, "path discovery"),
		strings.Contains(msg, "node not discovered"),
		strings.Contains(msg, "link establish"),
		strings.Contains(msg, "invalid node hash"),
		strings.Contains(msg, "invalid destination"):
		return KindConnectionFailed, errMsg
	case strings.Contains(msg, "context canceled"),
		strings.Contains(msg, "teardown"),
		strings.Contains(msg, "link closed"),
		strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "connection lost"):
		return KindConnectionLost, errMsg
	case strings.Contains(msg, "internal server error"),
		strings.Contains(msg, "internal error"),
		strings.Contains(msg, " 500"):
		return KindInternal, errMsg
	default:
		return KindInternal, errMsg
	}
}

func classifyBody(body []byte) (Kind, string) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return KindNotFound, "empty response"
	}
	lower := strings.ToLower(trimmed)
	if len(lower) > 512 {
		lower = lower[:512]
	}
	switch {
	case strings.Contains(lower, "404"), strings.Contains(lower, "not found"):
		return KindNotFound, trimmed
	case strings.Contains(lower, "500"), strings.Contains(lower, "internal server error"):
		return KindInternal, trimmed
	default:
		return KindUnknown, trimmed
	}
}

func IsCorruptError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "malformed") ||
		strings.Contains(msg, "corrupt") ||
		strings.Contains(msg, "not a database") ||
		strings.Contains(msg, "database disk image")
}

func ClassifyStorage(err error) Kind {
	if err == nil {
		return ""
	}
	if IsCorruptError(err) {
		return KindDatabaseCorrupt
	}
	if errors.Is(err, syscall.ENOSPC) || errors.Is(err, syscall.EROFS) || errors.Is(err, fs.ErrPermission) {
		return KindStorageFull
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "disk full"),
		strings.Contains(msg, "no space left"),
		strings.Contains(msg, "read-only file system"),
		strings.Contains(msg, "readonly database"),
		strings.Contains(msg, "unable to open database file"):
		return KindStorageFull
	default:
		return ""
	}
}

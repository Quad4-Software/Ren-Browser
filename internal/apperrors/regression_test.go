// SPDX-License-Identifier: MIT
package apperrors_test

import (
	"testing"

	"renbrowser/internal/apperrors"
)

func TestClassifyFetchRegressionMatrix(t *testing.T) {
	cases := []struct {
		msg  string
		kind apperrors.Kind
	}{
		{"reticulum not ready", apperrors.KindConnectionFailed},
		{"response too large: received 16 bytes (limit 8)", apperrors.KindPayloadTooLarge},
		{"empty response", apperrors.KindNotFound},
		{"no path to node", apperrors.KindConnectionFailed},
		{"context canceled", apperrors.KindConnectionLost},
		{"internal server error", apperrors.KindInternal},
		{"disk full", apperrors.KindInternal},
	}
	for _, tc := range cases {
		kind, _ := apperrors.ClassifyFetch(tc.msg, nil)
		if kind != tc.kind {
			t.Fatalf("%q: kind=%q want %q", tc.msg, kind, tc.kind)
		}
	}
}

func TestClassifyFetchBodyNotFound(t *testing.T) {
	kind, detail := apperrors.ClassifyFetch("", []byte("404 page not found"))
	if kind != apperrors.KindNotFound {
		t.Fatalf("kind=%q", kind)
	}
	if detail == "" {
		t.Fatal("expected detail")
	}
}

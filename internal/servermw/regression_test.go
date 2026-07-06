// SPDX-License-Identifier: MIT
package servermw_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"renbrowser/internal/servermw"
)

func TestRecoverHandlerPassesThroughSuccess(t *testing.T) {
	const body = "ok"
	h := servermw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(body))
	}), servermw.Options{})

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status=%d", rec.Code)
	}
	got, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("body=%q", got)
	}
}

func TestReadyHandlerDoesNotInterceptOtherPaths(t *testing.T) {
	h := servermw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), servermw.Options{
		ReadyCheck: func() bool { return false },
	})

	req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d", rec.Code)
	}
}

package servermw_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"renbrowser/internal/servermw"
)

func TestStripBasePath(t *testing.T) {
	called := ""
	h := servermw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}), servermw.Options{BasePath: "/ren"})

	req := httptest.NewRequest(http.MethodGet, "/ren/about", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if called != "/about" {
		t.Fatalf("path = %q", called)
	}
}

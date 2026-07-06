// SPDX-License-Identifier: MIT
package nomadnet

import (
	"strings"
	"testing"
)

func TestBuildRequestDataVars(t *testing.T) {
	out := buildRequestData(RequestData{Vars: map[string]string{"category": "general"}})
	if out == nil {
		t.Fatal("expected request data map")
	}
	if out["var_category"] != "general" {
		t.Fatalf("payload = %#v", out)
	}
}

func TestBuildRequestDataFormFields(t *testing.T) {
	out := buildRequestData(RequestData{Fields: map[string]string{"user": "alice"}})
	if out["field_user"] != "alice" {
		t.Fatalf("payload = %#v", out)
	}
}

func TestParseURLFieldPrefix(t *testing.T) {
	parsed, err := ParseURL("abb3ebcd03cb2388a838e70c001291f9:/page/form.mu`category=general|field.user=alice")
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if parsed.Request.Vars["category"] != "general" {
		t.Fatalf("vars = %#v", parsed.Request.Vars)
	}
	if parsed.Request.Fields["user"] != "alice" {
		t.Fatalf("fields = %#v", parsed.Request.Fields)
	}
}

func TestFormatURLWithRequestRoundTrip(t *testing.T) {
	req := RequestData{
		Vars:   map[string]string{"category": "general"},
		Fields: map[string]string{"user": "alice"},
	}
	raw := FormatURLWithRequest("abc", "/page/forum.mu", req)
	parsed, err := ParseURL(raw)
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if parsed.Request.Vars["category"] != "general" || parsed.Request.Fields["user"] != "alice" {
		t.Fatalf("request = %#v", parsed.Request)
	}
}

func TestRequestDataEmpty(t *testing.T) {
	var empty RequestData
	if !empty.Empty() {
		t.Fatal("expected empty request")
	}
	withVar := RequestData{Vars: map[string]string{"a": "1"}}
	if withVar.Empty() {
		t.Fatal("expected non-empty request")
	}
}

func TestRequestDataCacheKeySuffix(t *testing.T) {
	req := RequestData{
		Vars:   map[string]string{"category": "general"},
		Fields: map[string]string{"user": "alice"},
	}
	got := req.CacheKeySuffix()
	if got == "" || !strings.Contains(got, "category=general") || !strings.Contains(got, "field.user=alice") {
		t.Fatalf("suffix = %q", got)
	}
	var empty RequestData
	if empty.CacheKeySuffix() != "" {
		t.Fatal("expected empty suffix for empty request")
	}
}

func TestParseRequestPairsFieldPrefix(t *testing.T) {
	req := parseRequestPairs(map[string]string{
		"category":   "general",
		"field.user": "alice",
		"field.":     "ignored",
	})
	if req.Vars["category"] != "general" {
		t.Fatalf("vars = %#v", req.Vars)
	}
	if req.Fields["user"] != "alice" {
		t.Fatalf("fields = %#v", req.Fields)
	}
	if _, ok := req.Fields[""]; ok {
		t.Fatal("expected empty field name to be ignored")
	}
}

func TestParseRequestPairsCachesSuffix(t *testing.T) {
	req := parseRequestPairs(map[string]string{
		"category":   "general",
		"field.user": "alice",
	})
	if req.suffix == "" {
		t.Fatal("expected cached suffix")
	}
	if req.CacheKeySuffix() != req.suffix {
		t.Fatalf("suffix = %q cached = %q", req.CacheKeySuffix(), req.suffix)
	}
}

func TestBuildRequestDataEmpty(t *testing.T) {
	var empty RequestData
	if got := buildRequestData(empty); got != nil {
		t.Fatalf("expected nil payload, got %v", got)
	}
}

func TestFileNameFromMetadata(t *testing.T) {
	if got := fileNameFromMetadata(nil); got != "" {
		t.Fatalf("expected empty name for nil metadata, got %q", got)
	}
	if got := fileNameFromMetadata(map[string]any{"other": 1}); got != "" {
		t.Fatalf("expected empty name when no name key present, got %q", got)
	}
	if got := fileNameFromMetadata(map[string]any{"name": []byte("guide.zip")}); got != "guide.zip" {
		t.Fatalf("expected guide.zip for []byte name, got %q", got)
	}
	if got := fileNameFromMetadata(map[string]any{"name": "guide.zip"}); got != "guide.zip" {
		t.Fatalf("expected guide.zip for string name, got %q", got)
	}
}

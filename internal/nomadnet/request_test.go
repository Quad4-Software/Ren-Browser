package nomadnet

import (
	"strings"
	"testing"

	"quad4/msgpack/v5/pkg/msgpack"
)

func TestEncodeRequestDataMsgpackVars(t *testing.T) {
	raw := encodeRequestData(RequestData{Vars: map[string]string{"category": "general"}})
	if len(raw) == 0 {
		t.Fatal("expected msgpack payload")
	}
	var out map[string]string
	if err := msgpack.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["var_category"] != "general" {
		t.Fatalf("payload = %#v", out)
	}
}

func TestEncodeRequestDataFormFields(t *testing.T) {
	raw := encodeRequestData(RequestData{Fields: map[string]string{"user": "alice"}})
	var out map[string]string
	if err := msgpack.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
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

func TestEncodeRequestDataEmpty(t *testing.T) {
	var empty RequestData
	if got := encodeRequestData(empty); got != nil {
		t.Fatalf("expected nil payload, got %v", got)
	}
}

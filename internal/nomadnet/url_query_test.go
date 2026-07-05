package nomadnet

import "testing"

func TestParseURLMeshQueryFields(t *testing.T) {
	parsed, err := ParseURL("abb3ebcd03cb2388a838e70c001291f9:/page/form.mu?user=alice&action=go")
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if parsed.Path != "/page/form.mu" {
		t.Fatalf("path = %q", parsed.Path)
	}
	if parsed.Request.Vars["user"] != "alice" || parsed.Request.Vars["action"] != "go" {
		t.Fatalf("vars = %#v", parsed.Request.Vars)
	}
}

func TestParseURLMeshBacktickFields(t *testing.T) {
	parsed, err := ParseURL("abb3ebcd03cb2388a838e70c001291f9:/page/form.mu`user=alice|action=go")
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if parsed.Request.Vars["user"] != "alice" || parsed.Request.Vars["action"] != "go" {
		t.Fatalf("vars = %#v", parsed.Request.Vars)
	}
}

func TestParseURLMeshQueryAndBacktickFields(t *testing.T) {
	parsed, err := ParseURL("abb3ebcd03cb2388a838e70c001291f9:/page/form.mu?user=alice`action=go")
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if parsed.Request.Vars["user"] != "alice" || parsed.Request.Vars["action"] != "go" {
		t.Fatalf("vars = %#v", parsed.Request.Vars)
	}
}

func TestParseURLMeshQueryEscaping(t *testing.T) {
	parsed, err := ParseURL("abb3ebcd03cb2388a838e70c001291f9:/page/form.mu?user=alice%20bob&action=go%2Fhome")
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if parsed.Request.Vars["user"] != "alice bob" {
		t.Fatalf("user = %q", parsed.Request.Vars["user"])
	}
	if parsed.Request.Vars["action"] != "go/home" {
		t.Fatalf("action = %q", parsed.Request.Vars["action"])
	}
}

func TestParseURLMeshFilePath(t *testing.T) {
	parsed, err := ParseURL("abb3ebcd03cb2388a838e70c001291f9:/file/guide.zip")
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if parsed.Path != "/file/guide.zip" {
		t.Fatalf("path = %q", parsed.Path)
	}
}

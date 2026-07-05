package nomadnet_test

import (
	"testing"

	"renbrowser/internal/nomadnet"
)

func FuzzParseURLFields(f *testing.F) {
	f.Add("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu?a=1&b=2")
	f.Add("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu`a=1|b=2")
	f.Add("rns://ea6a715f814bdc37e56f80c34da6ad51/page/home.mu?q=test")

	f.Fuzz(func(t *testing.T, raw string) {
		parsed, err := nomadnet.ParseURL(raw)
		if err != nil {
			return
		}
		if parsed.NodeHash == "" || parsed.Path == "" {
			t.Fatalf("invalid parse for %q: %#v", raw, parsed)
		}
		for k := range parsed.Request.Vars {
			if k == "" {
				t.Fatalf("empty var key for %q", raw)
			}
		}
		for k := range parsed.Request.Fields {
			if k == "" {
				t.Fatalf("empty field key for %q", raw)
			}
		}
	})
}

// SPDX-License-Identifier: MIT
package content

import (
	"strings"
	"testing"
)

func TestIsolateNomadLinksRewritesRelative(t *testing.T) {
	in := `<a href="/page/other.mu">go</a>`
	out := isolateNomadLinks(in, "ABB3EBCD03CB2388A838E70C001291F9")
	if !strings.Contains(out, `data-nomad-url="abb3ebcd03cb2388a838e70c001291f9:/page/other.mu"`) {
		t.Fatalf("link not rewritten: %s", out)
	}
}

func TestIsolateNomadLinksNeutralizesExternal(t *testing.T) {
	in := `<a href="https://example.com" target="_blank">off mesh</a>`
	out := isolateNomadLinks(in, "abb3ebcd03cb2388a838e70c001291f9")
	if strings.Contains(out, `href="https://example.com"`) {
		t.Fatalf("external href not neutralized: %s", out)
	}
	if !strings.Contains(out, `href="#"`) {
		t.Fatalf("expected neutralized anchor: %s", out)
	}
}

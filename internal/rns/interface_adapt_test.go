// SPDX-License-Identifier: MIT
package rns

import (
	"strings"
	"testing"

	"quad4/reticulum-go/pkg/common"
)

func TestEffectiveInterfaceConfigPassthrough(t *testing.T) {
	cfg := &common.InterfaceConfig{
		Type:       "BackboneInterface",
		TargetHost: "rns.example.net",
		TargetPort: 4242,
		Enabled:    true,
	}
	got := EffectiveInterfaceConfig(cfg)
	if got != cfg {
		t.Fatal("expected same config pointer")
	}
	if got.Type != "BackboneInterface" {
		t.Fatalf("type = %q", got.Type)
	}
}

func TestParseInterfaceFragmentBackboneSnippet(t *testing.T) {
	snippet := "[[MichMesh]]\n  type = BackboneInterface\n  enabled = yes\n  remote = michmesh.example\n  target_port = 4242"
	ifaces, err := parseInterfaceFragment(snippet)
	if err != nil {
		t.Fatal(err)
	}
	cfg := ifaces["MichMesh"]
	if cfg == nil {
		t.Fatal("missing MichMesh interface")
	}
	if cfg.Type != "BackboneInterface" {
		t.Fatalf("type = %q, want BackboneInterface", cfg.Type)
	}
	if cfg.TargetHost != "michmesh.example" || cfg.TargetPort != 4242 {
		t.Fatalf("target = %s:%d", cfg.TargetHost, cfg.TargetPort)
	}
}

func TestFilterSeedableInterfacesIncludesBackbonePipeI2P(t *testing.T) {
	items := []CommunityInterface{
		{Name: "a", Type: "TCPClientInterface", Config: "[[a]]"},
		{Name: "b", Type: "backbone", TypeName: "BackboneInterface", Config: "[[b]]\n  type = BackboneInterface\n  remote = x.example\n  target_port = 4242"},
		{Name: "c", Type: "TCPServerInterface", Config: "[[c]]"},
		{Name: "d", Type: "pipe", TypeName: "PipeInterface", Config: "[[d]]\n  type = PipeInterface\n  command = rnsd"},
		{Name: "e", Type: "i2p", TypeName: "I2PInterface", Network: "i2p", Config: "[[e]]\n  type = I2PInterface"},
		{Name: "f", Type: "onion", Network: "onion", Config: "[[f]]"},
	}
	out := FilterSeedableInterfaces(items)
	if len(out) != 4 {
		t.Fatalf("len = %d, want 4", len(out))
	}
	names := make([]string, 0, len(out))
	for _, item := range out {
		names = append(names, item.Name)
	}
	joined := strings.Join(names, ",")
	if joined != "a,b,d,e" {
		t.Fatalf("names = %q", joined)
	}
}

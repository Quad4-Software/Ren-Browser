// SPDX-License-Identifier: MIT
package rns

import (
	"strings"
	"testing"

	"quad4/reticulum-go/pkg/common"
)

func TestBackboneToTCPClient(t *testing.T) {
	cfg := &common.InterfaceConfig{
		Type:       "BackboneInterface",
		TargetHost: "rns.example.net",
		TargetPort: 4242,
		Enabled:    true,
	}
	adapted, ok := backboneToTCPClient(cfg)
	if !ok {
		t.Fatal("expected conversion")
	}
	if adapted.Type != "TCPClientInterface" {
		t.Fatalf("type = %q", adapted.Type)
	}
	if adapted.TargetHost != "rns.example.net" || adapted.TargetPort != 4242 {
		t.Fatalf("target = %s:%d", adapted.TargetHost, adapted.TargetPort)
	}
}

func TestBackboneToTCPClientUsesAddressFallback(t *testing.T) {
	cfg := &common.InterfaceConfig{
		Type:    "BackboneInterface",
		Address: "mesh.example.net",
		Port:    7822,
	}
	adapted, ok := backboneToTCPClient(cfg)
	if !ok {
		t.Fatal("expected conversion")
	}
	if adapted.TargetHost != "mesh.example.net" || adapted.TargetPort != 7822 {
		t.Fatalf("target = %s:%d", adapted.TargetHost, adapted.TargetPort)
	}
}

func TestBackboneToTCPClientMissingHost(t *testing.T) {
	cfg := &common.InterfaceConfig{Type: "BackboneInterface", TargetPort: 4242}
	if _, ok := backboneToTCPClient(cfg); ok {
		t.Fatal("expected no conversion without host")
	}
}

func TestRewriteBackboneSnippetToTCP(t *testing.T) {
	snippet := "[[MichMesh]]\n  type = BackboneInterface\n  enabled = yes\n  remote = michmesh.example\n  target_port = 4242"
	out := rewriteBackboneSnippetToTCP(normalizeConfigSnippet(snippet))
	if strings.Contains(strings.ToLower(out), "backboneinterface") {
		t.Fatalf("backbone type remains: %q", out)
	}
	if !strings.Contains(out, "TCPClientInterface") {
		t.Fatalf("missing TCP client type: %q", out)
	}
	if !strings.Contains(out, "target_host = michmesh.example") {
		t.Fatalf("missing target_host: %q", out)
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
	if usesBackboneTCPFallback() {
		if cfg.Type != "TCPClientInterface" {
			t.Fatalf("type = %q, want TCPClientInterface on android", cfg.Type)
		}
		if cfg.TargetHost != "michmesh.example" || cfg.TargetPort != 4242 {
			t.Fatalf("target = %s:%d", cfg.TargetHost, cfg.TargetPort)
		}
	} else if cfg.Type != "BackboneInterface" {
		t.Fatalf("type = %q, want BackboneInterface on desktop", cfg.Type)
	}
}

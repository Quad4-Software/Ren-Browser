package rns

import (
	"fmt"
	"testing"

	"quad4/reticulum-go/pkg/common"
)

func TestPickSeedableCommunityInterfacesLimitsCount(t *testing.T) {
	items := make([]CommunityInterface, 0, 8)
	for i := range 8 {
		port := 4242 + i
		items = append(items, CommunityInterface{
			ID:       i + 1,
			Name:     fmt.Sprintf("node-%d", i+1),
			Type:     "TCPClientInterface",
			TypeName: "TCP Client",
			Network:  "clearnet",
			Host:     fmt.Sprintf("rns%d.example", i+1),
			Port:     &port,
			Status:   "online",
			Config:   fmt.Sprintf("[[node-%d]]\n  type = TCPClientInterface\n  enabled = yes\n  target_host = rns%d.example\n  target_port = %d\n", i+1, i+1, port),
		})
	}
	picked := PickSeedableCommunityInterfaces(items, 6)
	if len(picked) != 6 {
		t.Fatalf("PickSeedableCommunityInterfaces() len = %d, want 6", len(picked))
	}
}

func TestPickSeedableCommunityInterfacesSkipsI2PAndDuplicates(t *testing.T) {
	port := 4242
	items := []CommunityInterface{
		{
			Name: "good-a", Type: "TCPClientInterface", Network: "clearnet", Host: "a.example", Port: &port, Status: "online",
			Config: "[[good-a]]\n  type = TCPClientInterface\n  target_host = a.example\n  target_port = 4242\n",
		},
		{
			Name: "dup-a", Type: "backbone", TypeName: "BackboneInterface", Network: "clearnet", Host: "a.example", Port: &port, Status: "online",
			Config: "[[dup-a]]\n  type = BackboneInterface\n  remote = a.example\n  target_port = 4242\n",
		},
		{
			Name: "i2p", Type: "i2p", TypeName: "I2PInterface", Network: "i2p", Host: "x.b32.i2p", Status: "online",
			Config: "[[i2p]]\n  type = I2PInterface\n  peers = x.b32.i2p\n",
		},
		{
			Name: "offline", Type: "TCPClientInterface", Network: "clearnet", Host: "b.example", Port: &port, Status: "offline",
			Config: "[[offline]]\n  type = TCPClientInterface\n  target_host = b.example\n  target_port = 4242\n",
		},
		{
			Name: "good-b", Type: "TCPClientInterface", Network: "clearnet", Host: "b.example", Port: &port, Status: "online",
			Config: "[[good-b]]\n  type = TCPClientInterface\n  target_host = b.example\n  target_port = 4242\n",
		},
	}
	picked := PickSeedableCommunityInterfaces(items, 6)
	if len(picked) != 2 {
		t.Fatalf("len = %d, want 2 (deduped clearnet online only)", len(picked))
	}
	for _, item := range picked {
		if item.Network == "i2p" || item.Status == "offline" {
			t.Fatalf("unexpected item: %+v", item)
		}
	}
}

func TestConfigHasOutboundCommunityInterfaces(t *testing.T) {
	cfg := &common.ReticulumConfig{
		Interfaces: map[string]*common.InterfaceConfig{
			"Auto Discovery": {Type: "AutoInterface", Enabled: true},
		},
	}
	if ConfigHasOutboundCommunityInterfaces(cfg) {
		t.Fatal("AutoInterface-only config should not count as community outbound")
	}
	cfg.Interfaces["Mesh Hub"] = &common.InterfaceConfig{
		Type:    "TCPClientInterface",
		Enabled: true,
	}
	if !ConfigHasOutboundCommunityInterfaces(cfg) {
		t.Fatal("enabled TCP client should count as community outbound")
	}
}

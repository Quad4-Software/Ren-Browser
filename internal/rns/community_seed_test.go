package rns

import (
	"testing"

	"quad4/reticulum-go/pkg/common"
)

func TestPickSeedableCommunityInterfacesLimitsCount(t *testing.T) {
	items := make([]CommunityInterface, 0, 8)
	for i := range 8 {
		items = append(items, CommunityInterface{
			ID:       i + 1,
			Name:     "node",
			Type:     "TCPClientInterface",
			TypeName: "TCP Client",
			Config:   "[[TCP Client]]\n  enabled = yes\n",
		})
	}
	picked := PickSeedableCommunityInterfaces(items, 4)
	if len(picked) != 4 {
		t.Fatalf("PickSeedableCommunityInterfaces() len = %d, want 4", len(picked))
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

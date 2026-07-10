// SPDX-License-Identifier: MIT
package rns

import (
	"path/filepath"
	"testing"

	"quad4/reticulum-go/pkg/common"
)

func TestDeleteInterface(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	stack, err := NewStack(configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer stack.Stop()

	// Add a test interface
	stack.mu.Lock()
	if stack.cfg.Interfaces == nil {
		stack.cfg.Interfaces = make(map[string]*common.InterfaceConfig)
	}
	stack.cfg.Interfaces["test_iface"] = &common.InterfaceConfig{
		Type:    "AutoInterface",
		Enabled: true,
	}
	stack.mu.Unlock()

	// Verify it exists
	if _, ok := stack.cfg.Interfaces["test_iface"]; !ok {
		t.Fatal("expected interface to exist")
	}

	// Delete it
	if err := stack.DeleteInterface("test_iface"); err != nil {
		t.Fatalf("failed to delete interface: %v", err)
	}

	// Verify it's gone
	if _, ok := stack.cfg.Interfaces["test_iface"]; ok {
		t.Fatal("expected interface to be deleted")
	}

	// Try deleting non-existent interface
	if err := stack.DeleteInterface("non_existent"); err != errInterfaceNotFound {
		t.Fatalf("expected errInterfaceNotFound, got %v", err)
	}
}

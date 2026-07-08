package app

import (
	"path/filepath"
	"testing"

	"quad4/reticulum-go/pkg/common"
	"quad4/reticulum-go/pkg/reticulumconfig"

	"renbrowser/internal/rns"
)

func TestGetInitialSetupStateNeededForAutoOnlyConfig(t *testing.T) {
	svc := newTestBrowserService(t)
	state := svc.GetInitialSetupState()
	if !state.Needed {
		t.Fatal("fresh install should need initial setup")
	}
}

func TestGetInitialSetupStateSkipsWhenAlreadyComplete(t *testing.T) {
	svc := newTestBrowserService(t)
	svc.CompleteInitialSetup()
	state := svc.GetInitialSetupState()
	if state.Needed {
		t.Fatal("completed setup should not be needed again")
	}
}

func TestGetInitialSetupStateMigratesExistingCommunityConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config")
	cfg := &common.ReticulumConfig{
		ConfigPath: cfgPath,
		Interfaces: map[string]*common.InterfaceConfig{
			"Auto Discovery": {Name: "Auto Discovery", Type: "AutoInterface", Enabled: true},
			"Community Hub": {
				Name:    "Community Hub",
				Type:    "TCPClientInterface",
				Enabled: true,
			},
		},
	}
	if err := reticulumconfig.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	stack, err := rns.NewStack(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stack.Stop() })

	svc, err := NewBrowserServiceWithOptions(stack, nil, ServiceOptions{
		ProfilePath: filepath.Join(dir, "profile.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = svc.Store().Close() })

	state := svc.GetInitialSetupState()
	if state.Needed {
		t.Fatal("existing community interfaces should skip initial setup")
	}
	if !svc.GetBrowserPrefs().InitialSetupComplete {
		t.Fatal("migration should mark initial setup complete")
	}
}

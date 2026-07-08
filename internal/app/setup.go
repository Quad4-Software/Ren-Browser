// SPDX-License-Identifier: MIT
package app

import (
	"fmt"

	"quad4/reticulum-go/pkg/reticulumconfig"

	"renbrowser/internal/rns"
)

type InitialSetupState struct {
	Needed         bool `json:"needed"`
	SuggestedCount int  `json:"suggestedCount"`
}

func (s *BrowserService) GetInitialSetupState() InitialSetupState {
	state := InitialSetupState{
		SuggestedCount: rns.DefaultCommunityInterfaceCount,
	}
	if s.GetBrowserPrefs().InitialSetupComplete {
		return state
	}
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		state.Needed = true
		return state
	}
	cfg := stack.Config()
	if rns.ConfigHasOutboundCommunityInterfaces(cfg) {
		s.markInitialSetupComplete()
		return state
	}
	state.Needed = true
	return state
}

func (s *BrowserService) PreviewSuggestedCommunityInterfaces() ([]rns.CommunityInterface, error) {
	result, err := rns.FetchCommunityInterfaces(nil)
	if err != nil {
		return nil, err
	}
	return rns.PickSeedableCommunityInterfaces(result.Items, rns.DefaultCommunityInterfaceCount), nil
}

func (s *BrowserService) ApplySuggestedCommunityInterfaces() ([]string, error) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return nil, fmt.Errorf("reticulum not initialized")
	}
	path := stack.ConfigPath()
	cfg, err := reticulumconfig.LoadConfig(path)
	if err != nil {
		return nil, err
	}
	result, err := rns.FetchCommunityInterfaces(nil)
	if err != nil {
		return nil, err
	}
	added := rns.SeedCommunityInterfaces(cfg, result.Items, rns.DefaultCommunityInterfaceCount)
	if len(added) == 0 {
		return added, nil
	}
	if err := reticulumconfig.SaveConfig(cfg); err != nil {
		return nil, err
	}
	if err := stack.ApplyConfig(cfg); err != nil {
		return nil, err
	}
	s.log("info", "suggested community interfaces added", fmt.Sprintf("%v", added))
	if s.app != nil {
		s.app.Event.Emit("rns:status", "reload")
	}
	return added, nil
}

func (s *BrowserService) CompleteInitialSetup() {
	s.markInitialSetupComplete()
}

func (s *BrowserService) markInitialSetupComplete() {
	prefs := s.GetBrowserPrefs()
	if prefs.InitialSetupComplete {
		return
	}
	prefs.InitialSetupComplete = true
	s.SetBrowserPrefs(prefs)
}

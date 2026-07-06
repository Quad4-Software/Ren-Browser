// SPDX-License-Identifier: MIT
package app

import "renbrowser/internal/rns"

type CommunityFetchResult struct {
	Items      []rns.CommunityInterface `json:"items"`
	FromBundle bool                     `json:"fromBundle"`
}

func (s *BrowserService) FetchCommunityInterfaces() (CommunityFetchResult, error) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	installed := map[string]bool{}
	if stack != nil {
		installed = stack.InstalledInterfaceNames()
	}
	result, err := rns.FetchCommunityInterfaces(installed)
	if err != nil {
		return CommunityFetchResult{}, err
	}
	return CommunityFetchResult{
		Items:      result.Items,
		FromBundle: result.FromBundle,
	}, nil
}

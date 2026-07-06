// SPDX-License-Identifier: MIT

package app

// Shutdown stops mesh networking and requests application exit.
func (s *BrowserService) Shutdown() {
	_ = s.StopReticulum()
	s.mu.RLock()
	wailsApp := s.app
	s.mu.RUnlock()
	if wailsApp != nil {
		wailsApp.Quit()
	}
}

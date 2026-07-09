// SPDX-License-Identifier: MIT
package app

import "renbrowser/internal/nomadnet"

// PrepareForWake invalidates idle cached links and soft-stale transport paths
// after the UI returns from background/suspend. Call before the next page
// reload so Fetch does not reuse zombie routes from before sleep.
func (s *BrowserService) PrepareForWake() nomadnet.WakePrepResult {
	if s == nil {
		return nomadnet.WakePrepResult{}
	}
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil || stack.Browser() == nil {
		return nomadnet.WakePrepResult{}
	}
	return stack.Browser().PrepareForWake()
}

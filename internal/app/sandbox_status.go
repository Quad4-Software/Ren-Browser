// SPDX-License-Identifier: MIT
package app

import "renbrowser/internal/sandbox"

// SandboxStatus describes process-level filesystem sandboxing.
type SandboxStatus = sandbox.Status

func (s *BrowserService) GetSandboxStatus() SandboxStatus {
	return sandbox.CurrentStatus()
}

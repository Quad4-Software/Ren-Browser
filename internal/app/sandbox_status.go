// SPDX-License-Identifier: MIT
package app

import (
	"renbrowser/internal/runtimeenv"
	"renbrowser/internal/sandbox"
)

// SandboxStatus describes process-level filesystem sandboxing.
type SandboxStatus = sandbox.Status

func (s *BrowserService) GetSandboxStatus() SandboxStatus {
	st := sandbox.CurrentStatus()
	env := runtimeenv.Detect()
	st.InFlatpak = env.InFlatpak
	st.InAppImage = env.InAppImage
	st.InContainer = env.InContainer
	st.ContainerRuntime = env.ContainerRuntime
	st.WebKitSandbox = env.WebKitSandbox
	st.WebKitSandboxNote = env.WebKitSandboxNote
	st.OnAndroid = env.OnAndroid
	return st
}

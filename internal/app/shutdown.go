// SPDX-License-Identifier: MIT
package app

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func (s *BrowserService) Shutdown() {
	s.shutdown(true)
}

func (s *BrowserService) ServiceShutdown(ctx context.Context, options application.ServiceOptions) error {
	s.shutdown(false)
	return nil
}

func (s *BrowserService) shutdown(quitApp bool) {
	s.shutdownOnce.Do(func() {
		s.mu.Lock()
		s.shuttingDown = true
		downloads := s.downloads
		stack := s.stack
		plugins := s.plugins
		st := s.store
		wailsApp := s.app
		s.mu.Unlock()

		_, _ = s.capturePrimaryWindowState()

		if downloads != nil {
			downloads.shutdownInFlight(downloadInterruptedText)
			s.persistDownloadRecovery(downloads.list())
		}
		_ = s.StopReticulum()
		if stack != nil && stack.Browser() != nil {
			stack.Browser().Close()
		}
		if plugins != nil {
			_ = plugins.Close()
		}
		if st != nil {
			_ = st.Close()
		}
		if quitApp && wailsApp != nil {
			wailsApp.Quit()
		}
	})
}

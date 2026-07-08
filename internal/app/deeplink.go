// SPDX-License-Identifier: MIT
package app

import (
	"renbrowser/internal/deeplink"
)

// DeepLinkEvent is emitted when the OS opens the app with a URL.
const DeepLinkEvent = deeplink.EventName

// HandleDeepLink accepts an OS-provided launch URL, unwraps it, queues it, and
// emits app:deeplink for the frontend.
func (s *BrowserService) HandleDeepLink(rawURL string) string {
	target, ok := deeplink.Unwrap(rawURL)
	if !ok {
		return ""
	}
	deeplink.Enqueue(target)
	s.emitDeepLink(target)
	return target
}

// TakePendingDeepLink returns and clears any queued deeplink target.
func (s *BrowserService) TakePendingDeepLink() string {
	return deeplink.TakePending()
}

func (s *BrowserService) emitDeepLink(target string) {
	if s == nil || s.app == nil || target == "" {
		return
	}
	s.app.Event.Emit(DeepLinkEvent, target)
}

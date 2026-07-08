// SPDX-License-Identifier: MIT
package bootstrap

import (
	"os"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"renbrowser/internal/app"
	"renbrowser/internal/brand"
	"renbrowser/internal/deeplink"
)

func wireDeepLinks(wailsApp *application.App, browserSvc *app.BrowserService) {
	if wailsApp == nil || browserSvc == nil {
		return
	}

	deeplink.SetHandler(func(target string) {
		if target == "" {
			return
		}
		wailsApp.Event.Emit(app.DeepLinkEvent, target)
	})

	_ = wailsApp.Event.OnApplicationEvent(events.Common.ApplicationLaunchedWithUrl, func(event *application.ApplicationEvent) {
		if event == nil || event.Context() == nil {
			return
		}
		browserSvc.HandleDeepLink(event.Context().URL())
	})

	// Cold-start argv fallback when the platform event has not fired yet.
	if deeplink.PeekPending() == "" {
		if target, ok := deeplink.ExtractFromArgs(os.Args[1:]); ok {
			deeplink.Enqueue(target)
		}
	}
}

func singleInstanceOptions(browserSvc *app.BrowserService) *application.SingleInstanceOptions {
	return &application.SingleInstanceOptions{
		UniqueID: brand.BundleID,
		OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
			if browserSvc == nil {
				return
			}
			if target, ok := deeplink.ExtractFromArgs(data.Args); ok {
				browserSvc.HandleDeepLink(target)
				return
			}
			for _, arg := range data.Args {
				arg = strings.TrimSpace(arg)
				if arg == "" || strings.HasPrefix(arg, "-") {
					continue
				}
				lower := strings.ToLower(arg)
				if strings.Contains(arg, "://") || strings.HasPrefix(lower, "renbrowser:") {
					browserSvc.HandleDeepLink(arg)
					return
				}
			}
		},
	}
}

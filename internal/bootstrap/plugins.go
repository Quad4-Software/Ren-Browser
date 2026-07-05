// SPDX-License-Identifier: MIT
package bootstrap

import (
	"fmt"

	"renbrowser/internal/app"
	"renbrowser/internal/plugins"
)

func setupPlugins(browserSvc *app.BrowserService) (*plugins.Manager, *app.PluginHost, error) {
	st := browserSvc.Store()
	if st == nil {
		return nil, nil, fmt.Errorf("store unavailable")
	}
	manager := plugins.NewManager(st)
	browserSvc.SetPluginManager(manager)
	if err := manager.LoadAll(); err != nil {
		return nil, nil, err
	}
	return manager, app.NewPluginHost(manager), nil
}

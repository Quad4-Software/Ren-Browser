//go:build server

// SPDX-License-Identifier: MIT

package bootstrap

import (
	"embed"

	"renbrowser/internal/app"
	"renbrowser/internal/config"
	"renbrowser/internal/rns"
	"renbrowser/internal/serverauth"
)

func New(embedded embed.FS, cfg config.Runtime) (*App, error) {
	HandleResetIfNeeded(cfg)

	_, loader, err := openAssetsAndLoader(embedded, cfg)
	if err != nil {
		return nil, err
	}

	stack, err := rns.NewStack(cfg.ReticulumConfig)
	if err != nil {
		_ = loader.Close()
		return nil, err
	}

	browserSvc, err := app.NewBrowserServiceWithOptions(stack, nil, app.ServiceOptions{
		ProfileName:   cfg.Profile,
		PublicMode:    cfg.PublicMode,
		ResetWindow:   cfg.ResetWindow,
		ExportProfile: cfg.ExportProfile,
		ImportProfile: cfg.ImportProfile,
	})
	if err != nil {
		_ = loader.Close()
		return nil, err
	}

	guard, err := serverauth.Configure(browserSvc.Store().DB(), cfg)
	if err != nil {
		_ = loader.Close()
		return nil, err
	}

	pluginMgr, pluginHost, err := setupPlugins(browserSvc)
	if err != nil {
		_ = loader.Close()
		return nil, err
	}

	extra := WailsServerExtra{}
	if guard != nil && guard.Enabled() {
		mw := guard.Middleware()
		extra = WailsServerExtra{
			ServerWrap: mw,
		}
	}

	wailsApp := newWailsApp(browserSvc, pluginHost, pluginMgr, loader, cfg, extra)
	browserSvc.SetApp(wailsApp)
	if pluginMgr != nil {
		pluginMgr.SetApp(wailsApp)
	}

	return &App{
		Wails:      wailsApp,
		Service:    browserSvc,
		PluginHost: pluginHost,
		PluginMgr:  pluginMgr,
		Loader:     loader,
	}, nil
}

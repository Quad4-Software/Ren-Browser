//go:build !android
// SPDX-License-Identifier: MIT

package bootstrap

import (
	"embed"

	"renbrowser/internal/app"
	"renbrowser/internal/config"
	"renbrowser/internal/rns"
)

func New(embedded embed.FS, cfg config.Runtime) (*App, error) {
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

	wailsApp := newWailsApp(browserSvc, loader, cfg)
	browserSvc.SetApp(wailsApp)

	return &App{
		Wails:   wailsApp,
		Service: browserSvc,
		Loader:  loader,
	}, nil
}

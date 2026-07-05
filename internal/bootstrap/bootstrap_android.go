//go:build android

package bootstrap

import (
	"embed"
	"log"

	"renbrowser/internal/app"
	"renbrowser/internal/config"
	"renbrowser/internal/paths"
	"renbrowser/internal/rns"
)

func New(embedded embed.FS, cfg config.Runtime) (*App, error) {
	paths.InitAndroid()

	_, loader, err := openAssetsAndLoader(embedded, cfg)
	if err != nil {
		return nil, err
	}

	browserSvc, err := app.NewBrowserServiceWithOptions(nil, nil, app.ServiceOptions{
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

	go func() {
		stack, err := rns.NewStack(cfg.ReticulumConfig)
		if err != nil {
			log.Printf("reticulum stack: %v", err)
			return
		}
		browserSvc.AttachStack(stack)
		if err := browserSvc.StartReticulum(); err != nil {
			log.Printf("reticulum start: %v", err)
		}
	}()

	return &App{
		Wails:   wailsApp,
		Service: browserSvc,
		Loader:  loader,
	}, nil
}

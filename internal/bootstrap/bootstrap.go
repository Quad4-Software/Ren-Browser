package bootstrap

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/wailsapp/wails/v3/pkg/application"

	"renbrowser/internal/app"
	"renbrowser/internal/assets"
	"renbrowser/internal/config"
	"renbrowser/internal/nomadnet"
	"renbrowser/internal/servermw"
)

type App struct {
	Wails   *application.App
	Service *app.BrowserService
	Loader  *assets.Loader
}

func openAssetsAndLoader(embedded embed.FS, cfg config.Runtime) (fs.FS, *assets.Loader, error) {
	assetFS, err := openAssets(embedded, cfg)
	if err != nil {
		return nil, nil, err
	}

	loader, err := assets.New(assets.Config{
		Embedded: assetFS,
		Dir:      cfg.AssetsDir,
		ZipPath:  cfg.AssetsZip,
	})
	if err != nil {
		return nil, nil, err
	}
	return assetFS, loader, nil
}

func newWailsApp(browserSvc *app.BrowserService, loader *assets.Loader, cfg config.Runtime) *application.App {
	registerEvents()

	handler := servermw.Wrap(loader.Handler(), servermw.Options{
		TrustProxy: cfg.TrustProxy,
		BasePath:   cfg.BasePath,
	})

	return application.New(application.Options{
		Name:        "Ren Browser",
		Description: "Reticulum browser for NomadNet pages",
		Services: []application.Service{
			application.NewService(browserSvc),
		},
		Assets: application.AssetOptions{
			Handler: handler,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: false,
		},
		Server: application.ServerOptions{
			Host: cfg.ServerHost,
			Port: cfg.ServerPort,
		},
	})
}

func openAssets(embedded embed.FS, cfg config.Runtime) (fs.FS, error) {
	if cfg.AssetsDir != "" || cfg.AssetsZip != "" {
		return nil, nil
	}
	sub, err := assets.SubEmbedded(embedded, "frontend/dist")
	if err != nil {
		return nil, fmt.Errorf("embedded assets: %w", err)
	}
	return sub, nil
}

func registerEvents() {
	application.RegisterEvent[string]("rns:status")
	application.RegisterEvent[app.PageResponse]("page:loaded")
	application.RegisterEvent[app.PageResponse]("page:error")
	application.RegisterEvent[[]nomadnet.Node]("node:discovered")
	application.RegisterEvent[app.ThemeSettings]("theme:changed")
	application.RegisterEvent[string]("dev:log")
	application.RegisterEvent[app.NetworkEntry]("network:entry")
	application.RegisterEvent[app.WindowChrome]("window:chrome")
	application.RegisterEvent[app.RuntimeConfig]("runtime:config")
	application.RegisterEvent[app.StoreHealth]("store:health")
}

func AssetHandlerForServer(loader *assets.Loader, cfg config.Runtime) http.Handler {
	return servermw.Wrap(loader.Handler(), servermw.Options{
		TrustProxy: cfg.TrustProxy,
		BasePath:   cfg.BasePath,
	})
}

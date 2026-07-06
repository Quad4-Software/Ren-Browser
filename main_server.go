//go:build server

// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"fmt"
	"os"

	"renbrowser/internal/app"
	"renbrowser/internal/bootstrap"
	"renbrowser/internal/brand"
	"renbrowser/internal/buildinfo"
	"renbrowser/internal/config"
	"renbrowser/internal/serverlog"
)

//go:embed all:frontend/dist
var embeddedAssets embed.FS

func main() {
	cfg := config.ParseFlags()
	if cfg.ServerHost == "" {
		cfg.ServerHost = "0.0.0.0"
	}
	if cfg.ServerPort == 0 {
		cfg.ServerPort = 8080
	}

	appBundle, err := bootstrap.New(embeddedAssets, cfg)
	if err != nil {
		serverlog.Init()
		serverlog.Error("server bootstrap failed", err)
		os.Exit(1)
	}
	defer appBundle.Loader.Close()

	logServerStartup(appBundle.Service, cfg)

	if err := appBundle.Wails.Run(); err != nil {
		serverlog.Error("server stopped", err)
		os.Exit(1)
	}
}

func logServerStartup(svc *app.BrowserService, cfg config.Runtime) {
	serverlog.Init()
	serverlog.Banner(brand.DisplayName+" server", brand.Version, buildinfo.BuildLabel())

	configPath := cfg.ReticulumConfig
	if configPath == "" {
		configPath = svc.ConfigPath()
	}
	serverlog.Field("config", configPath)
	serverlog.Field("profile", brand.ProfileDBPath(cfg.Profile))
	if cfg.PublicMode {
		serverlog.Field("mode", "public (browser-side storage)")
	}
	if cfg.BasePath != "" {
		serverlog.Field("base path", cfg.BasePath)
	}
	if cfg.TrustProxy {
		serverlog.Field("proxy", "trust X-Forwarded-* headers")
	}

	host := cfg.ServerHost
	if host == "" {
		host = "0.0.0.0"
	}
	port := cfg.ServerPort
	if port == 0 {
		port = 8080
	}
	serverlog.Field("listen", fmt.Sprintf("http://%s:%d/", host, port))
	if host == "0.0.0.0" || host == "::" {
		serverlog.Field("open", localBrowserURL(port, cfg.BasePath))
	}

	if err := svc.StartReticulum(); err != nil {
		serverlog.Error("Reticulum failed to start", err)
	} else {
		logMeshStatus(svc)
	}

	serverlog.Ready(localBrowserURL(port, cfg.BasePath))
}

func logMeshStatus(svc *app.BrowserService) {
	st := svc.GetStatus()
	if st.InterfaceCount == 0 {
		serverlog.Warn("no Reticulum interfaces configured")
		return
	}
	serverlog.OK(fmt.Sprintf("Reticulum started (%d/%d interfaces online, %d nodes announced)",
		st.InterfacesOnline, st.InterfaceCount, st.NodeCount))

	for _, iface := range svc.ListInterfaces() {
		state := "offline"
		switch {
		case !iface.Enabled:
			state = "disabled"
		case iface.Online:
			state = "online"
		}
		serverlog.Info(fmt.Sprintf("interface %s (%s) %s", iface.Name, iface.Type, state))
	}
}

func localBrowserURL(port int, basePath string) string {
	path := normalizeBasePath(basePath)
	return fmt.Sprintf("http://127.0.0.1:%d%s", port, path)
}

func normalizeBasePath(basePath string) string {
	path := basePath
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if path == "" {
		return "/"
	}
	return path
}

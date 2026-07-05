//go:build android && !server
// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"log"

	"renbrowser/internal/bootstrap"
	"renbrowser/internal/config"
)

//go:embed all:frontend/dist
var embeddedAssets embed.FS

func main() {
	cfg := config.ParseFlags()

	appBundle, err := bootstrap.New(embeddedAssets, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer appBundle.Loader.Close()

	prefs := appBundle.Service.GetBrowserPrefs()
	frameless := !prefs.NativeTitlebar
	windowOpts := appBundle.Service.InitialWindowOptions(frameless, cfg.ResetWindow)

	_ = appBundle.Wails.Window.NewWithOptions(windowOpts)

	if err := appBundle.Wails.Run(); err != nil {
		log.Fatal(err)
	}
}

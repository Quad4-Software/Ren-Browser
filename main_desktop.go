//go:build !server && !android

package main

import (
	"embed"
	"log"
	"net/http"
	"os"

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

	win := appBundle.Wails.Window.NewWithOptions(windowOpts)
	appBundle.Service.AttachWindowPersistence(win)

	go func() {
		if err := appBundle.Service.StartReticulum(); err != nil {
			log.Printf("reticulum start: %v", err)
		}
	}()

	if os.Getenv("REN_BROWSER_ASSET_PROBE") == "1" {
		log.Printf("asset source: %s", appBundle.Loader.Kind())
		_, _ = http.Get("http://127.0.0.1")
	}

	if err := appBundle.Wails.Run(); err != nil {
		log.Fatal(err)
	}
}

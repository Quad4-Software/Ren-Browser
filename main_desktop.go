//go:build !server && !android

// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"renbrowser/internal/bootstrap"
	"renbrowser/internal/config"
)

//go:embed all:frontend/dist
var embeddedAssets embed.FS

func main() {
	cfg := config.ParseFlags()
	relocateForAppImage(&cfg)

	appBundle, err := bootstrap.New(embeddedAssets, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer appBundle.Loader.Close()

	prefs := appBundle.Service.GetBrowserPrefs()
	frameless := !prefs.NativeTitlebar
	windowOpts := appBundle.Service.InitialWindowOptions(frameless, cfg.ResetWindow)

	_ = appBundle.Wails.Window.NewWithOptions(windowOpts)

	maybeCaptureDesktopScreenshot()

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

// relocateForAppImage works around a WebKitGTK/AppImage relocation issue.
//
// build/linux/appimage/bundle-webkitgtk.sh patches the WebKitGTK shared
// libraries so the compiled-in helper process path (normally an absolute
// path such as /usr/lib/x86_64-linux-gnu/webkitgtk-6.0/WebKitNetworkProcess)
// becomes a same-length relative path (././lib/x86_64-linux-gnu/webkitgtk-6.0/
// WebKitNetworkProcess). glib resolves that path against the process's
// current working directory when it spawns WebKitNetworkProcess, not against
// the AppImage mount point, so launching the AppImage from any directory
// other than its own mount root makes the spawn fail with "No such file or
// directory". The bundled helpers live under $APPDIR/usr, so chdir there to
// match. See https://github.com/tauri-apps/tauri/issues/5292.
func relocateForAppImage(cfg *config.Runtime) {
	appDir := os.Getenv("APPDIR")
	if appDir == "" || os.Getenv("APPIMAGE") == "" {
		return
	}
	usrDir := filepath.Join(appDir, "usr")
	if info, err := os.Stat(usrDir); err != nil || !info.IsDir() {
		return
	}
	absolutize(&cfg.ReticulumConfig)
	absolutize(&cfg.AssetsDir)
	absolutize(&cfg.AssetsZip)
	absolutize(&cfg.ExportProfile)
	absolutize(&cfg.ImportProfile)
	_ = os.Chdir(usrDir)
}

func absolutize(path *string) {
	if *path == "" {
		return
	}
	if abs, err := filepath.Abs(*path); err == nil {
		*path = abs
	}
}

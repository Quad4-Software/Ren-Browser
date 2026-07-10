//go:build !server && !android

// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	_ "renbrowser/internal/webkit"
	"strings"

	"renbrowser/internal/app"
	"renbrowser/internal/bootstrap"
	"renbrowser/internal/brand"
	"renbrowser/internal/buildinfo"
	"renbrowser/internal/config"
	"renbrowser/internal/safe"
	"renbrowser/internal/sandbox"
)

//go:embed all:frontend/dist
var embeddedAssets embed.FS

func main() {
	cfg := config.ParseFlags()
	if cfg.Version {
		buildinfo.PrintVersion(brand.DisplayName)
		os.Exit(0)
	}
	relocateForAppImage(&cfg)
	opts := sandbox.OptionsFromRuntime(cfg)
	opts.ServerMode = false
	sandbox.Apply(opts)

	appBundle, err := bootstrap.New(embeddedAssets, cfg)
	if err != nil {
		handleFatalError(err)
	}
	defer appBundle.Loader.Close()

	if cfg.SelfCheck {
		log.Println("Running self-check diagnostics...")
		// Start Reticulum stack temporarily to allow full self-check
		if err := appBundle.Service.StartReticulum(); err != nil {
			log.Printf("Self-check warning: reticulum start failed: %v", err)
		}
		res := appBundle.Service.RunSelfCheck()
		_ = appBundle.Service.StopReticulum() // clean up stack
		logSelfCheckResult(res)
		if res.AllPassed {
			log.Println("Self-check PASSED!")
			os.Exit(0)
		}
		log.Println("Self-check FAILED")
		os.Exit(1)
	}

	if cfg.Headless {
		log.Println("Headless mode active, exiting successfully.")
		os.Exit(0)
	}

	if cfg.NativeTitlebar {
		prefs := appBundle.Service.GetBrowserPrefs()
		prefs.NativeTitlebar = true
		appBundle.Service.SetBrowserPrefs(prefs)
	}

	prefs := appBundle.Service.GetBrowserPrefs()
	frameless := !prefs.NativeTitlebar
	windowOpts := appBundle.Service.InitialWindowOptions(frameless, cfg.ResetWindow)

	_ = appBundle.Wails.Window.NewWithOptions(windowOpts)

	maybeCaptureDesktopScreenshot()

	safe.Go("reticulum-start", func() {
		if err := appBundle.Service.StartReticulum(); err != nil {
			log.Printf("reticulum start: %v", err)
		}
	})

	if os.Getenv("REN_BROWSER_ASSET_PROBE") == "1" {
		log.Printf("asset source: %s", appBundle.Loader.Kind())
		_, _ = http.Get("http://127.0.0.1")
	}

	if err := appBundle.Wails.Run(); err != nil {
		handleFatalError(err)
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
	if os.Getenv("APPIMAGE") == "" {
		return
	}
	appDir, err := filepath.Abs(filepath.Clean(os.Getenv("APPDIR")))
	if err != nil || appDir == "" {
		return
	}
	usrDir := filepath.Join(appDir, "usr")
	usrAbs, err := filepath.Abs(usrDir)
	if err != nil {
		return
	}
	rel, err := filepath.Rel(appDir, usrAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return
	}
	if info, err := os.Stat(usrAbs); err != nil || !info.IsDir() {
		return
	}
	absolutize(&cfg.ReticulumConfig)
	absolutize(&cfg.AssetsDir)
	absolutize(&cfg.AssetsZip)
	absolutize(&cfg.ExportProfile)
	absolutize(&cfg.ImportProfile)
	_ = os.Chdir(usrAbs)

	// WebKit's BubblewrapLauncher spawns its web process sandbox (bwrap plus
	// an xdg-dbus-proxy helper) with hard-coded, same-relocation-trick paths
	// like the one patched above. bwrap then bind-mounts a fixed set of
	// standard host paths for its own sandboxed view and has no idea our
	// helpers live under a transient AppImage mount point, so it can't see
	// them there and the dbus-proxy exec fails. There is no supported way to
	// add that mount from outside WebKit, so fall back to WebKit's documented
	// escape hatch for nested-sandbox conflicts. Flatpak sets the same env in
	// finish-args (and via internal/webkit) for the GNOME 50 / glibc 2.42+
	// BubblewrapLauncher crash.
	_ = os.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "1")
}

func absolutize(path *string) {
	if *path == "" {
		return
	}
	if abs, err := filepath.Abs(*path); err == nil {
		*path = abs
	}
}

func logSelfCheckResult(res app.SelfCheckResult) {
	logCheck := func(name string, st app.CheckStatus) {
		status := "PASS"
		if !st.Passed {
			status = "FAIL"
		}
		if st.Reason != "" {
			log.Printf("  %-12s %s — %s", name, status, st.Reason)
			return
		}
		log.Printf("  %-12s %s", name, status)
	}
	log.Printf("Self-check results (mesh=%v):", res.MeshEnabled)
	logCheck("stack", res.StackUp)
	logCheck("config", res.ConfigGood)
	logCheck("db", res.DBGood)
	logCheck("readwrite", res.ReadWriteGood)
	logCheck("downloads", res.DownloadsGood)
	logCheck("interfaces", res.Interfaces)
	logCheck("discovery", res.Discovery)
	logCheck("page", res.PageFetch)
}

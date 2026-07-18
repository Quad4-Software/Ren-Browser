// SPDX-License-Identifier: MIT
package sandbox

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"

	"renbrowser/internal/brand"
	"renbrowser/internal/config"
	"renbrowser/internal/paths"
	"renbrowser/internal/rns"
)

// OptionsFromRuntime builds sandbox path rules from startup configuration.
func OptionsFromRuntime(cfg config.Runtime) Options {
	retDir := rns.DefaultConfigDir()
	retConfig := cfg.ReticulumConfig
	if retConfig == "" {
		retConfig = filepath.Join(retDir, "config")
	}

	opts := Options{
		NoLandlock:      cfg.NoLandlock,
		ForceLandlock:   cfg.Landlock,
		NoSeccomp:       cfg.NoSeccomp,
		ForceSeccomp:    cfg.Seccomp,
		DataDir:         paths.Join(brand.DataDirName),
		ReticulumDir:    retDir,
		ReticulumConfig: retConfig,
		PluginsDir:      paths.Join(brand.DataDirName, "plugins"),
		AssetsDir:       cfg.AssetsDir,
		AssetsZip:       cfg.AssetsZip,
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		opts.DownloadDir = filepath.Join(home, "Downloads")
	}
	if dir := stringsTrim(xdg.UserDirs.Download); dir != "" {
		opts.DownloadDir = dir
	}
	if cache := stringsTrim(xdg.CacheHome); cache != "" {
		opts.ExtraReadPaths = append(opts.ExtraReadPaths, cache)
	}
	if runtimeDir := stringsTrim(os.Getenv("XDG_RUNTIME_DIR")); runtimeDir != "" {
		opts.ExtraReadPaths = append(opts.ExtraReadPaths, runtimeDir)
	}
	if appImage := stringsTrim(os.Getenv("APPIMAGE")); appImage != "" {
		opts.ExtraReadPaths = append(opts.ExtraReadPaths, appImage, filepath.Dir(appImage))
	}
	if appDir := stringsTrim(os.Getenv("APPDIR")); appDir != "" {
		opts.ExtraReadPaths = append(opts.ExtraReadPaths, appDir)
	}
	if exe, err := os.Executable(); err == nil && exe != "" {
		opts.ExtraReadPaths = append(opts.ExtraReadPaths, filepath.Dir(exe))
	}

	return opts
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}

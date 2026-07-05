// SPDX-License-Identifier: MIT
package app

import (
	"runtime"
	"strings"

	"renbrowser/internal/brand"
	"renbrowser/internal/buildinfo"
	"renbrowser/internal/content"
	"renbrowser/internal/store"
)

type AboutInfo struct {
	AppName         string `json:"appName"`
	Version         string `json:"version"`
	Build           string `json:"build"`
	License         string `json:"license"`
	GoVersion       string `json:"goVersion"`
	OS              string `json:"os"`
	Arch            string `json:"arch"`
	WailsVersion    string `json:"wailsVersion"`
	ReticulumConfig string `json:"reticulumConfig"`
	DataPath        string `json:"dataPath"`
}

func (s *BrowserService) GetAboutInfo() AboutInfo {
	return AboutInfo{
		AppName:         brand.DisplayName,
		Version:         buildinfo.Version,
		Build:           buildinfo.BuildLabel(),
		License:         content.ProjectLicense,
		GoVersion:       runtime.Version(),
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		WailsVersion:    brand.WailsVersion,
		ReticulumConfig: s.stack.ConfigPath(),
		DataPath:        store.DefaultPath(),
	}
}

func isAboutURL(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "about", "about:":
		return true
	default:
		return false
	}
}

func (s *BrowserService) aboutPage(pushHistory bool) PageResponse {
	info := s.GetAboutInfo()
	html := content.RenderAbout(content.AboutInfo{
		AppName:         info.AppName,
		Version:         info.Version,
		Build:           info.Build,
		License:         info.License,
		GoVersion:       info.GoVersion,
		OS:              info.OS,
		Arch:            info.Arch,
		WailsVersion:    info.WailsVersion,
		ReticulumConfig: info.ReticulumConfig,
		DataPath:        info.DataPath,
	})
	resp := PageResponse{
		URL:         "about:",
		Path:        "/about",
		ContentType: "about",
		HTML:        html,
		Raw:         html,
	}
	if pushHistory {
		s.pushHistory("about:")
		_ = s.store.AddHistory("about:", "About", "")
	}
	s.setLastPage(resp)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
	}
	return resp
}

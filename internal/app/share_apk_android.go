//go:build android

// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"renbrowser/internal/qr"
)

type preparedApk struct {
	OK       bool   `json:"ok"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Filename string `json:"filename"`
	Version  string `json:"version"`
	Error    string `json:"error,omitempty"`
}

var apkShareServer struct {
	sync.Mutex
	srv *http.Server
	url string
}

func prepareShareApk() (preparedApk, error) {
	raw, ok := androidBridgeString("prepareShareApkJson")
	if !ok || strings.TrimSpace(raw) == "" {
		return preparedApk{}, errors.New("apk share unavailable")
	}
	var out preparedApk
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return preparedApk{}, err
	}
	if !out.OK {
		if out.Error != "" {
			return out, errors.New(out.Error)
		}
		return out, errors.New("failed to prepare apk")
	}
	return out, nil
}

func localWiFiIP() string {
	raw, ok := androidBridgeString("getLocalIpJson")
	if !ok {
		return ""
	}
	var v struct {
		IP string `json:"ip"`
	}
	if json.Unmarshal([]byte(raw), &v) != nil {
		return ""
	}
	return strings.TrimSpace(v.IP)
}

func apkShareQRDataURL(url string) string {
	if strings.TrimSpace(url) == "" {
		return ""
	}
	dataURL, err := qr.DataURL(url, 4)
	if err != nil {
		return ""
	}
	return dataURL
}

func (s *BrowserService) GetApkShareInfo() ApkShareInfo {
	apk, err := prepareShareApk()
	if err != nil {
		return ApkShareInfo{Error: err.Error()}
	}
	return ApkShareInfo{
		Available: true,
		Version:   apk.Version,
		Size:      apk.Size,
		Filename:  apk.Filename,
	}
}

func (s *BrowserService) ShareApk() error {
	androidBridgeVoid("sharePreparedApk")
	return nil
}

func (s *BrowserService) StartApkShareServer() ApkShareSession {
	apkShareServer.Lock()
	defer apkShareServer.Unlock()

	if apkShareServer.srv != nil {
		url := apkShareServer.url
		return ApkShareSession{Active: true, URL: url, QRDataURL: apkShareQRDataURL(url)}
	}

	apk, err := prepareShareApk()
	if err != nil {
		return ApkShareSession{Error: err.Error()}
	}

	ip := localWiFiIP()
	if ip == "" {
		return ApkShareSession{Error: "no local network address found; connect to Wi-Fi first"}
	}

	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return ApkShareSession{Error: err.Error()}
	}

	port := ln.Addr().(*net.TCPAddr).Port
	filename := apk.Filename
	path := apk.Path

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/"+filename {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(path)
		if err != nil {
			http.Error(w, "apk unavailable", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "application/vnd.android.package-archive")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		_, _ = io.Copy(w, f)
	})

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln)
	}()

	url := fmt.Sprintf("http://%s:%d/%s", ip, port, filename)
	apkShareServer.srv = srv
	apkShareServer.url = url
	return ApkShareSession{Active: true, URL: url, QRDataURL: apkShareQRDataURL(url)}
}

func (s *BrowserService) StopApkShareServer() {
	apkShareServer.Lock()
	defer apkShareServer.Unlock()
	if apkShareServer.srv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = apkShareServer.srv.Shutdown(ctx)
	apkShareServer.srv = nil
	apkShareServer.url = ""
}

func (s *BrowserService) GetApkShareSession() ApkShareSession {
	apkShareServer.Lock()
	defer apkShareServer.Unlock()
	if apkShareServer.srv == nil {
		return ApkShareSession{}
	}
	url := apkShareServer.url
	return ApkShareSession{Active: true, URL: url, QRDataURL: apkShareQRDataURL(url)}
}

//go:build !android

// SPDX-License-Identifier: MIT

package app

import "errors"

var errApkShareUnsupported = errors.New("apk sharing is only available on android")

func (s *BrowserService) GetApkShareInfo() ApkShareInfo {
	return ApkShareInfo{Error: errApkShareUnsupported.Error()}
}

func (s *BrowserService) ShareApk() error {
	return errApkShareUnsupported
}

func (s *BrowserService) StartApkShareServer() ApkShareSession {
	return ApkShareSession{Error: errApkShareUnsupported.Error()}
}

func (s *BrowserService) StopApkShareServer() {}

func (s *BrowserService) GetApkShareSession() ApkShareSession {
	return ApkShareSession{}
}

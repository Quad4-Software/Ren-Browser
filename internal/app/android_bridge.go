//go:build android

// SPDX-License-Identifier: MIT

package app

import _ "unsafe"

//go:linkname androidBridgeString github.com/wailsapp/wails/v3/pkg/application.androidBridgeString
func androidBridgeString(method string) (string, bool)

//go:linkname androidBridgeStringString github.com/wailsapp/wails/v3/pkg/application.androidBridgeStringString
func androidBridgeStringString(method string, arg string) (string, bool)

//go:linkname androidBridgeVoid github.com/wailsapp/wails/v3/pkg/application.androidBridgeVoid
func androidBridgeVoid(method string)

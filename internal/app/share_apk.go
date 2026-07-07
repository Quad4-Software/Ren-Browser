// SPDX-License-Identifier: MIT

package app

// ApkShareInfo describes the installed Android package available for sharing.
type ApkShareInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Size      int64  `json:"size"`
	Filename  string `json:"filename"`
	Error     string `json:"error,omitempty"`
}

// ApkShareSession describes an active local HTTP APK download session.
type ApkShareSession struct {
	Active    bool   `json:"active"`
	URL       string `json:"url"`
	QRDataURL string `json:"qrDataURL,omitempty"`
	Error     string `json:"error,omitempty"`
}

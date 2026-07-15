// SPDX-License-Identifier: MIT
// Healthcheck for distroless images with no shell or curl.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	port := os.Getenv("WAILS_SERVER_PORT")
	if port == "" {
		port = os.Getenv("REN_BROWSER_PORT")
	}
	if port == "" {
		port = "8080"
	}

	base := strings.Trim(strings.TrimSpace(os.Getenv("REN_BROWSER_BASE_PATH")), "/")
	path := "/health"
	if base != "" {
		path = "/" + base + "/health"
	}

	for _, c := range port {
		if c < '0' || c > '9' {
			os.Exit(1)
		}
	}

	// Host is always loopback. Port is digits-only from container env.
	url := fmt.Sprintf("http://127.0.0.1:%s%s", port, path)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url) // #nosec G704 -- fixed 127.0.0.1 health probe
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}

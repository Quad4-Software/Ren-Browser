//go:build server

package main

import (
	"embed"
	"log"

	"renbrowser/internal/bootstrap"
	"renbrowser/internal/config"
)

//go:embed all:frontend/dist
var embeddedAssets embed.FS

func main() {
	cfg := config.ParseFlags()
	if cfg.ServerHost == "" {
		cfg.ServerHost = "0.0.0.0"
	}
	if cfg.ServerPort == 0 {
		cfg.ServerPort = 8080
	}

	appBundle, err := bootstrap.New(embeddedAssets, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer appBundle.Loader.Close()

	go func() {
		if err := appBundle.Service.StartReticulum(); err != nil {
			log.Printf("reticulum start: %v", err)
		}
	}()

	log.Printf("Ren Browser server listening on %s:%d", cfg.ServerHost, cfg.ServerPort)
	if err := appBundle.Wails.Run(); err != nil {
		log.Fatal(err)
	}
}

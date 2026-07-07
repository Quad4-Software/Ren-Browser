// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"renbrowser/internal/plugins"
)

func main() {
	root := flag.String("root", "..", "extension root directory")
	wasmPath := flag.String("wasm", "../translator.wasm", "wasm module to bundle")
	outPath := flag.String("out", "../renbrowser.micron-translator.wasm", "bundled output path")
	flag.Parse()

	manifestRaw, err := os.ReadFile(filepath.Join(*root, plugins.ManifestFileName)) // #nosec G304 -- build tool path
	if err != nil {
		fmt.Fprintf(os.Stderr, "bundle: read manifest: %v\n", err)
		os.Exit(1)
	}
	var manifest plugins.Manifest
	if err := json.Unmarshal(manifestRaw, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "bundle: parse manifest: %v\n", err)
		os.Exit(1)
	}
	wasmData, err := os.ReadFile(*wasmPath) // #nosec G304 -- build tool path
	if err != nil {
		fmt.Fprintf(os.Stderr, "bundle: read wasm: %v\n", err)
		os.Exit(1)
	}
	files := map[string]string{}
	for _, name := range []string{"main.js", "settings.js"} {
		raw, err := os.ReadFile(filepath.Join(*root, name)) // #nosec G304 -- build tool path
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			fmt.Fprintf(os.Stderr, "bundle: read %s: %v\n", name, err)
			os.Exit(1)
		}
		files[name] = string(raw)
	}
	localeEntries, err := os.ReadDir(filepath.Join(*root, "locales"))
	if err == nil {
		for _, entry := range localeEntries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(*root, "locales", entry.Name())) // #nosec G304 -- build tool path
			if err != nil {
				fmt.Fprintf(os.Stderr, "bundle: read locales/%s: %v\n", entry.Name(), err)
				os.Exit(1)
			}
			files["locales/"+entry.Name()] = string(raw)
		}
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "bundle: read locales: %v\n", err)
		os.Exit(1)
	}
	bundled, err := plugins.BundleWasm(wasmData, manifest, files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bundle: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*outPath, bundled, 0o600); err != nil { // #nosec G703 -- build tool output path from flags
		fmt.Fprintf(os.Stderr, "bundle: write output: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("bundle: wrote %s\n", *outPath)
}

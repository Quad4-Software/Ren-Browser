// SPDX-License-Identifier: MIT
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"renbrowser/internal/plugins"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "sign":
		if err := runSign(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "pluginsign: %v\n", err)
			os.Exit(1)
		}
	case "verify":
		if err := runVerify(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "pluginsign: %v\n", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  pluginsign sign -identity PATH -zip PATH
  pluginsign sign -identity PATH -dir PATH
  pluginsign sign -identity PATH -wasm PATH [-output PATH]
  pluginsign verify -zip PATH
  pluginsign verify -dir PATH
  pluginsign verify -wasm PATH

Signing uses the Reticulum rnid utility to create .rsg signatures and embeds
them as renbrowser.plugin.rsg (zip/dir) or a renbrowser.signature wasm section.
`)
}

func runSign(args []string) error {
	fs := flag.NewFlagSet("sign", flag.ContinueOnError)
	identity := fs.String("identity", "", "path to Reticulum identity (.rid)")
	zipPath := fs.String("zip", "", "plugin zip to sign")
	dirPath := fs.String("dir", "", "plugin directory to sign")
	wasmPath := fs.String("wasm", "", "plugin wasm bundle to sign")
	output := fs.String("output", "", "output wasm path (defaults to in-place)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*identity) == "" {
		return fmt.Errorf("-identity is required")
	}
	switch {
	case *zipPath != "":
		return signZip(*identity, *zipPath)
	case *dirPath != "":
		return signDir(*identity, *dirPath)
	case *wasmPath != "":
		return signWasm(*identity, *wasmPath, *output)
	default:
		return fmt.Errorf("one of -zip, -dir, or -wasm is required")
	}
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	zipPath := fs.String("zip", "", "plugin zip to verify")
	dirPath := fs.String("dir", "", "plugin directory to verify")
	wasmPath := fs.String("wasm", "", "plugin wasm bundle to verify")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var info plugins.SignatureInfo
	switch {
	case *zipPath != "":
		info = plugins.VerifyZipSignature(*zipPath)
	case *dirPath != "":
		info = plugins.VerifyDirSignature(*dirPath)
	case *wasmPath != "":
		data, err := os.ReadFile(*wasmPath) // #nosec G304 -- path from CLI arg
		if err != nil {
			return err
		}
		info = plugins.VerifyWasmSignature(data)
	default:
		return fmt.Errorf("one of -zip, -dir, or -wasm is required")
	}
	return printSignature(info)
}

func signZip(identity, zipPath string) error {
	reader, err := zip.OpenReader(zipPath) // #nosec G304 -- path from CLI arg
	if err != nil {
		return err
	}
	defer reader.Close()
	payload, err := plugins.CanonicalZipPayload(reader.File)
	if err != nil {
		return err
	}
	rsg, err := signPayload(identity, payload)
	if err != nil {
		return err
	}
	return plugins.EmbedSignatureInZip(zipPath, rsg)
}

func signDir(identity, dir string) error {
	payload, err := plugins.CanonicalDirPayload(dir)
	if err != nil {
		return err
	}
	rsg, err := signPayload(identity, payload)
	if err != nil {
		return err
	}
	return plugins.WriteDirSignature(dir, rsg)
}

func signWasm(identity, wasmPath, output string) error {
	data, err := os.ReadFile(wasmPath) // #nosec G304 -- path from CLI arg
	if err != nil {
		return err
	}
	payload, err := plugins.WasmPayloadWithoutSignature(data)
	if err != nil {
		return err
	}
	rsg, err := signPayload(identity, payload)
	if err != nil {
		return err
	}
	signed, err := plugins.AppendWasmSignature(data, rsg)
	if err != nil {
		return err
	}
	outPath := wasmPath
	if strings.TrimSpace(output) != "" {
		outPath = output
	}
	return os.WriteFile(outPath, signed, 0o600) // #nosec G703 -- output path from CLI flag or source wasm path
}

func signPayload(identity string, payload []byte) ([]byte, error) {
	tmp, err := os.CreateTemp("", "renplugin-sign-*")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close() // #nosec G104 -- cleanup after write failure
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	rnid := os.Getenv("RNID")
	if rnid == "" {
		rnid = "rnid"
	}
	cmd := exec.Command(rnid, "-i", identity, "-s", tmpPath) // #nosec G204 G702 -- explicit signing tool and operator-controlled paths
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("rnid sign failed: %w", err)
	}
	rsgPath := tmpPath + ".rsg"
	rsg, err := os.ReadFile(rsgPath) // #nosec G304 -- rnid output beside temp payload
	if err != nil {
		return nil, err
	}
	defer os.Remove(rsgPath)
	return rsg, nil
}

func printSignature(info plugins.SignatureInfo) error {
	if !info.Present {
		fmt.Println("unsigned")
		return nil
	}
	if !info.Valid {
		if info.Error != "" {
			return fmt.Errorf("invalid signature: %s", info.Error)
		}
		return fmt.Errorf("invalid signature")
	}
	if info.Trusted && info.SignerName != "" {
		fmt.Printf("valid signer=%s name=%s trusted=yes\n", info.Signer, info.SignerName)
	} else {
		fmt.Printf("valid signer=%s trusted=no\n", info.Signer)
	}
	return nil
}

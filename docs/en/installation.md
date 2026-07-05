# Installation

This page covers pre-built downloads, Docker, and building from source.

## Pre-built downloads (recommended)

Get the latest release for your system from [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases).

| Platform | File | Notes |
|----------|------|-------|
| Linux x86_64 | `renbrowser-linux-amd64.AppImage` | `chmod +x` then run. A plain binary is also included. |
| Linux ARM64 | `renbrowser-linux-arm64.AppImage` | Same steps as x86_64. |
| Windows | `renbrowser-windows-amd64.exe` | Run directly. No installer required. |
| macOS | `renbrowser-macos-universal.zip` | Unzip and open `renbrowser.app`. |
| Server (Linux x86_64) | `renbrowser-server-linux-amd64` | Headless binary for Docker or self-hosting. |
| Android | `renbrowser.apk` | When the release pipeline includes it. |

Each release ships `SHA256SUMS.txt` so you can verify downloads. See [Security](security.md).

### Verify a download (Linux or macOS)

```sh
sha256sum -c SHA256SUMS.txt
```

Check only the file you downloaded if the sums file lists many assets.

## Docker or Podman (server mode)

Official image: `ghcr.io/quad4-software/renbrowser`

Mount your Reticulum config so the container can join the mesh:

```sh
docker run --rm -p 8080:8080 \
  -v "$HOME/.reticulum-go:/root/.reticulum-go:ro" \
  ghcr.io/quad4-software/renbrowser:latest
```

Open `http://localhost:8080` in any browser on the same machine.

Build and run from this repo:

```sh
task build:docker
task run:docker
```

The server image has **no login screen**. Only expose it on networks you trust. See [Server mode](server-mode.md) and [Security](security.md).

## Build from source

For contributors or platforms without a pre-built package.

### Requirements

- [Go](https://go.dev/) 1.26 or newer
- [Node.js](https://nodejs.org/) 22+ and [pnpm](https://pnpm.io/) 11+
- [Task](https://taskfile.dev/) (recommended)
- Reticulum config at `~/.reticulum-go/` (or set `REN_BROWSER_CONFIG`)

### Basic build

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task build
./bin/renbrowser
```

Go modules pull Quad4 dependencies from GitHub automatically.

### Platform-specific builds

```sh
task build:windows
task build:darwin
task build:android      # physical device (arm64)
task build:android:emu  # emulator (host ABI)
```

### Installers and packages

```sh
task package                  # current OS
task package:linux:appimage   # Linux AppImage
task package:darwin:universal # macOS universal
task package:windows          # Windows NSIS installer
```

### Android SDK

Android builds need the [Android SDK](https://developer.android.com/studio) (API 34, NDK r26+). Set `ANDROID_HOME` and run `task android:install:deps` if the build reports missing tools.

## Server binary from source

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

See [Server mode](server-mode.md) for environment variables and deployment notes.

## After install

1. Confirm Reticulum config is in place ([Reticulum setup](reticulum-setup.md))
2. Launch the app and open `about:`
3. Read [Using the browser](using-the-browser.md)

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
| Server (Linux ARM64) | `renbrowser-server-linux-arm64` | Raspberry Pi 3/4/5 and other 64-bit ARM boards. |
| Server (Linux ARMv6) | `renbrowser-server-linux-armv6` | Raspberry Pi Zero W and other 32-bit ARMv6 devices. |
| Server (FreeBSD) | `renbrowser-server-freebsd-amd64`, `renbrowser-server-freebsd-arm64` | Headless on FreeBSD. |
| Server (OpenBSD / NetBSD) | `renbrowser-server-openbsd-amd64`, `renbrowser-server-netbsd-amd64` | Headless on BSD. |
| Android | `renbrowser.apk` | When the release pipeline includes it. |

Each release ships `SHA256SUMS.txt` so you can verify downloads. See [Security](security.md).

### Verify a download (Linux or macOS)

```sh
sha256sum -c SHA256SUMS.txt
```

Check only the file you downloaded if the sums file lists many assets.

### System requirements

| Package | What you need on the host |
|---------|---------------------------|
| **Linux AppImage** | Bundles GTK 4, WebKitGTK 6, and other libraries. No separate WebKit install. Some distros need FUSE or `APPIMAGE_EXTRACT_AND_RUN=1`. |
| **Linux Flatpak** | Flatpak plus the `org.gnome.Platform` runtime (GTK 4 and WebKitGTK 6). |
| **Linux plain binary** | GTK 4 and WebKitGTK 6.0 at runtime (for example on Debian/Ubuntu 24.04+, Fedora, or Arch). |
| **Windows `.exe`** | [Microsoft Edge WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/). Usually on Windows 10/11. The NSIS installer can install it; the portable `.exe` does not. |
| **macOS `.app`** | Recent macOS with system WebKit (no extra runtime). |
| **Android APK** | Android 5.0+ (API 21+). |
| **Server binary / Docker** | No desktop GUI stack. Use any browser on the host for the UI. Release server builds: Linux amd64/arm64/armv6, FreeBSD amd64/arm64, OpenBSD/NetBSD amd64. |

## Docker or Podman (server mode)

Official image: `ghcr.io/quad4-software/renbrowser`

Mount your Reticulum config and profile data so the container can join the mesh. The image runs as a non-root user, so pass your host UID/GID:

```sh
docker run --rm -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

The same flags work with `podman run`. On Podman you can use `--userns=keep-id` instead of `--user "$(id -u):$(id -g)"`. If SELinux blocks the bind mount, add `:Z` to the volume flags.

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

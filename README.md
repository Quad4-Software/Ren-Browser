# Ren Browser

Desktop browser for [NomadNet](https://github.com/markqvist/NomadNet) pages over [Reticulum](https://reticulum.network/). Built with Wails v3, Go, and Svelte 5.

## Requirements

- Go 1.26+
- [pnpm](https://pnpm.io/) 11+
- [Task](https://taskfile.dev/) (recommended)
- Reticulum config at `~/.reticulum-go/`

Go dependencies (`quad4/reticulum-go`, `micron-parser-go`, and related Quad4 modules) resolve from [Quad4-Software](https://github.com/Quad4-Software) via `replace` directives in `go.mod`. No sibling-repo checkout is required.

To bump Quad4 deps to the latest `master` commits:

```sh
go get github.com/Quad4-Software/Reticulum-Go@master
go get github.com/Quad4-Software/Micron-Parser-Go@master
go mod tidy
```

## Development

```sh
task dev
```

## Build

```sh
task build          # current platform
task build:windows  # Windows
task build:darwin   # macOS
task build:android      # Android debug APK (arm64, physical devices)
task build:android:emu  # Android debug APK (host ABI, emulator)
```

Package installers:

```sh
task package
task package:windows
task package:darwin
task package:android   # signed release APK (arm64 + x86_64)
```

Android requires the [Android SDK](https://developer.android.com/studio) with platform-tools, platform API 34, build-tools, and NDK r26+. Set `ANDROID_HOME` (or `ANDROID_SDK_ROOT`). Install deps with `task android:install:deps` if needed.

## Test

```sh
task check
task test:interop   # live Reticulum network test
```

## Server mode

Headless HTTP server for Docker and reverse-proxy deployments:

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Environment variables (also loadable from `.env`):

- `WAILS_SERVER_HOST` / `REN_BROWSER_HOST`
- `WAILS_SERVER_PORT` / `REN_BROWSER_PORT`
- `REN_BROWSER_TRUST_PROXY` — trust `X-Forwarded-*` headers
- `REN_BROWSER_BASE_PATH` — subpath prefix behind a reverse proxy
- `REN_BROWSER_CONFIG` — Reticulum config path

Docker:

```sh
task build:docker
task run:docker
```

## Runtime data

Application state is stored in `~/.renbrowser/renbrowser.db` (SQLite, WAL). Legacy `~/.renbrowser/state.json` is migrated on first run.

## Flags

- `--config` — Reticulum config file path
- `--assets-dir` — serve frontend from a directory instead of embedded assets
- `--assets-zip` — serve frontend from a zip archive
- `--host` / `--port` — HTTP bind address (server mode)
- `--trust-proxy` — honor reverse-proxy forwarded headers
- `--base-path` — URL prefix when served behind a reverse proxy

## Layout

- `main_desktop.go` / `main_server.go` — Wails desktop and headless entry points
- `internal/` — Reticulum, NomadNet, rendering, SQLite store
- `frontend/` — Svelte UI
- `build/` — platform packaging and Wails config

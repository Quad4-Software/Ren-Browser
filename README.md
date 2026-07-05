# Ren Browser

A modern browser for Reticulum Network.

This project is under heavy active development, please wait until v1.0 for stability and to be more user friendly.

## Documentation

Guides by language in [docs/](docs/):

| Language | Folder |
|----------|--------|
| English | [docs/en/](docs/en/) |
| Russian | [docs/ru/](docs/ru/) |
| Spanish | [docs/es/](docs/es/) |
| German | [docs/de/](docs/de/) |

Start with [Getting started](docs/en/getting-started.md) in your language.

## Install

### Pre-built downloads

Grab the latest release for your system from [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases).

### Docker or Podman

The published server image is `ghcr.io/quad4-software/renbrowser-server`.

Mount your Reticulum config so the container can join the mesh:

```sh
docker run --rm -p 8080:8080 \
  -v "$HOME/.reticulum-go:/data/reticulum" \
  -e REN_BROWSER_CONFIG=/data/reticulum/config \
  ghcr.io/quad4-software/renbrowser-server:latest
```

Then open `http://localhost:8080` in any browser on the same machine.

For a custom build from this repo:

```sh
task build:docker
task run:docker
```

The server image currently has **no login screen**. Only expose it on networks you trust, or put it behind a reverse proxy with access controls. See [SECURITY.md](SECURITY.md).

### Build from source

For contributors or platforms without a pre-built package.

**You will need:**

- [Go](https://go.dev/) 1.26 or newer
- [Node.js](https://nodejs.org/) 22+ and [pnpm](https://pnpm.io/) 11+
- [Task](https://taskfile.dev/) (recommended)
- Reticulum config at `~/.reticulum-go/` (or set `REN_BROWSER_CONFIG`)

**Steps:**

```sh
git clone https://github.com/Quad4-Software/Ren-Browser.git
cd Ren-Browser
task build
./bin/renbrowser
```

Go modules pull Quad4 dependencies from GitHub automatically; no extra repos to clone.

Platform-specific builds:

```sh
task build:windows
task build:darwin
task build:android      # physical device (arm64)
task build:android:emu  # emulator (host ABI)
```

Installers (AppImage, `.app` bundle, etc.):

```sh
task package
```

Android builds need the [Android SDK](https://developer.android.com/studio) (API 34, NDK r26+). Set `ANDROID_HOME` and run `task android:install:deps` if the build complains about missing tools.

## Using the app

- **Address bar** — enter a NomadNet destination or use built-in schemes (`about:`, `settings:`, etc.).
- **Discovery** — find nodes announced on your Reticulum interfaces.
- **Settings** — manage interfaces, themes, extensions, and profile data.
- **Data** — bookmarks, history, and tabs are stored in `~/.renbrowser/renbrowser.db`. Older `state.json` files are migrated on first launch.

Type `license` in the address bar to read the in-app license text.

## Extensions

Install extensions from **Settings → Extensions** (zip or folder), or unpack into `~/.renbrowser/plugins/<id>/` with a `renbrowser.plugin.json` manifest.

An example extension lives in `extensions/hello-extension/`. Extension authors work with permissions (storage, navigation, network, and related caps) declared in the manifest. See `internal/plugins/manifest.go` for the full schema.

## Server mode

Run Ren Browser as a web app without the desktop shell — useful for homelab servers, Docker, or a machine that already runs Reticulum.

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Common environment variables (also readable from a `.env` file in the working directory):

| Variable | Purpose |
|----------|---------|
| `WAILS_SERVER_HOST` / `REN_BROWSER_HOST` | Bind address |
| `WAILS_SERVER_PORT` / `REN_BROWSER_PORT` | Port (default `8080`) |
| `REN_BROWSER_CONFIG` | Path to Reticulum config |
| `REN_BROWSER_TRUST_PROXY` | Trust `X-Forwarded-*` from a reverse proxy |
| `REN_BROWSER_BASE_PATH` | URL prefix when served under a subpath |

Use `--public-mode` to keep favorites, history, and tabs in the browser (`localStorage`) instead of the server database.

## Development

```sh
task dev
```

Run the full quality gate before sending changes:

```sh
task check
task test:interop   # optional; needs a live Reticulum network
```

## Project layout

| Path | Contents |
|------|----------|
| `main_desktop.go` / `main_server.go` | Desktop and headless entry points |
| `internal/` | Reticulum, NomadNet, rendering, SQLite store, plugins |
| `frontend/` | Svelte 5 UI |
| `build/` | Packaging and platform tooling |

## Contributing

Patches and guidance: [CONTRIBUTING.md](CONTRIBUTING.md)

Security reports: [SECURITY.md](SECURITY.md)

Legal and licensing questions: [LEGAL.md](LEGAL.md)

## License

Ren Browser is released under the [MIT License](LICENSE).

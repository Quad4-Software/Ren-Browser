# Server mode

Server mode runs Ren Browser as a web app without the desktop shell. You access it from another browser at an HTTP URL.

## When to use server mode

- Homelab or VPS that already runs Reticulum
- Docker deployments
- Shared machine where you prefer not to install the desktop app
- Access from tablets or phones on your LAN

## Quick start

```sh
task build:server
./bin/renbrowser-server --host 0.0.0.0 --port 8080
```

Open `http://localhost:8080` (or your host IP) in Firefox, Chromium, or Safari.

## Docker

Published image:

```
ghcr.io/quad4-software/renbrowser:latest
```

Example run (create the host directories first so Docker does not create them as root):

```sh
mkdir -p "$HOME/.reticulum-go" "$HOME/.renbrowser"
docker run --rm --name renbrowser -p 8080:8080 \
  --user "$(id -u):$(id -g)" \
  -e HOME=/data \
  -v "$HOME/.reticulum-go:/data/.reticulum-go" \
  -v "$HOME/.renbrowser:/data/.renbrowser" \
  -e REN_BROWSER_CONFIG=/data/.reticulum-go/config \
  ghcr.io/quad4-software/renbrowser:latest
```

Do not mount the Reticulum directory read-only; the mesh needs to update storage beside the config file. See [Reticulum setup](reticulum-setup.md#server-and-docker) for mount details and Podman notes.

Build locally:

```sh
task build:docker
task run:docker
```

## Command-line flags

| Flag | Purpose |
|------|---------|
| `--host` | Bind address (default `0.0.0.0` in server build) |
| `--port` | HTTP port (default `8080`) |
| `--config` | Reticulum config path |
| `--trust-proxy` | Trust `X-Forwarded-*` from a reverse proxy |
| `--base-path` | URL prefix when served under a subpath |
| `--public-mode` | Store favorites, history, and tabs in browser `localStorage` instead of server SQLite |
| `--profile` | Named profile database |
| `--import-profile` / `--export-profile` | Profile JSON at startup |

## Environment variables

The server reads a `.env` file in the working directory. Variables already set in the environment are not overwritten.

| Variable | Purpose |
|----------|---------|
| `WAILS_SERVER_HOST` / `REN_BROWSER_HOST` | Bind address |
| `WAILS_SERVER_PORT` / `REN_BROWSER_PORT` | Port |
| `REN_BROWSER_CONFIG` / `RETICULUM_CONFIG` | Reticulum config |
| `REN_BROWSER_TRUST_PROXY` | `true` / `1` / `yes` to enable trust proxy |
| `REN_BROWSER_BASE_PATH` | Subpath prefix |
| `REN_BROWSER_PUBLIC_MODE` | Public mode toggle |
| `REN_BROWSER_PROFILE` | Profile name |
| `REN_BROWSER_IMPORT_PROFILE` | Import path at startup |
| `REN_BROWSER_EXPORT_PROFILE` | Export path at startup |

## Public mode

Without `--public-mode`, the server keeps tabs, history, and favorites in its SQLite database on the server disk. Every client sharing that instance sees the same data.

With `--public-mode`, those items live in each browser's `localStorage`. Use this when many people hit one server and should not share one profile.

## Reverse proxy

Typical nginx or Caddy setup:

1. Terminate TLS at the proxy
2. Proxy to `127.0.0.1:8080`
3. Pass `X-Forwarded-Proto` and `X-Forwarded-Host`
4. Start Ren Browser with `--trust-proxy`
5. Set `--base-path` if the app is not at the domain root

Header `X-RenBrowser-Base-Path` is recognized when trust proxy is on.

## No built-in authentication

Anyone who can reach the HTTP port can use the browser and trigger Reticulum traffic. Do not expose port 8080 to the public internet without:

- Firewall rules
- VPN
- Reverse proxy with auth
- Or all of the above

Read [Security](security.md) before you publish a server.

## Asset overrides (advanced)

For development you can serve frontend files from disk or zip instead of embedded assets:

- `--assets-dir path`
- `--assets-zip path`

Environment: `REN_BROWSER_ASSETS_DIR`, `REN_BROWSER_ASSETS_ZIP`.

## Next steps

- [Data and profiles](data-and-profiles.md) for SQLite vs public mode
- [Security](security.md) for deployment hardening
- [Installation](installation.md) for release binaries

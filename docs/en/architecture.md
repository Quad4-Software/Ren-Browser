# Architecture

High-level view of how Ren Browser is structured.

## Overview

```
┌─────────────────────────────────────────────────────────┐
│  Svelte 5 frontend (Wails webview or browser in server) │
│  Tabs, chrome, Micron viewer, panels, settings          │
└───────────────────────┬─────────────────────────────────┘
                        │ Wails bindings / HTTP API
┌───────────────────────▼─────────────────────────────────┐
│  internal/app : BrowserService                          │
│  Navigation, tabs, history, settings, plugin bridge     │
└───────┬─────────────┬─────────────┬─────────────────────┘
        │             │             │
        ▼             ▼             ▼
   internal/rns   internal/store  internal/plugins
   Reticulum      SQLite          Extensions
        │
        ▼
   internal/nomadnet : LXMF page fetch for announced NomadNet nodes
        │
        ▼
   internal/micron : Markup parse and HTML render
```

## Entry points

| File | Build tag | Role |
|------|-----------|------|
| `main_desktop.go` | `!server && !android` | Wails window, embedded `frontend/dist` |
| `main_server.go` | `server` | HTTP server, same embedded assets |
| Android main | `android` | Mobile shell |

`internal/bootstrap` wires config, store, plugins, and Wails app together.

## Frontend

- **Framework:** Svelte 5 with Vite
- **Bindings:** Generated under `frontend/bindings/renbrowser/`
- **Main UI:** `frontend/src/App.svelte` orchestrates chrome and panels
- **Components:** `frontend/src/lib/components/` (tab bar, discovery, settings, etc.)
- **Browser logic:** `frontend/src/lib/browser/` (URLs, keybinds, errors)

Micron rendering can use WASM parsers managed by `MicronWasmManager` with SRI verification.

## Backend services

### BrowserService (`internal/app`)

Central API for the UI:

- Navigate URLs and manage tab state
- Expose discovery, history, favorites
- Load and save preferences
- Bridge to plugin host

### Reticulum stack (`internal/rns`)

Wraps `quad4/reticulum-go`:

- Start and stop transport
- Report interface stats
- Hot reload config from Settings

### Page fetch (`internal/nomadnet`)

Fetches remote `.mu` and related content over LXMF and Reticulum. Discovery labels nodes as NomadNet when their announces match. Ren Browser does not use NomadNet client libraries.

### Store (`internal/store` + `internal/db`)

SQLite persistence with migration from legacy `state.json`.

### Plugins (`internal/plugins`)

- Manifest validation
- Permission enforcement
- Builtin schemes (`about:`, `license:`, `editor:`)
- JS and WASM plugin runtimes

## Content and rendering

| Package | Role |
|---------|------|
| `internal/content` | Static pages (about, license) |
| `internal/micron` | Micron to HTML |
| `internal/micronwasm` | WASM parser integration |
| `internal/cache` | Page cache helpers |

## Server middleware

`internal/servermw` handles base path headers and proxy-aware URL building in server mode.

## Configuration

`internal/config` parses flags, `.env`, and `REN_BROWSER_*` variables into a `Runtime` struct used by bootstrap.

## Brand and paths

`internal/brand` (generated from `build/brand.yml`) defines stable names:

- Data dir `.renbrowser`
- DB file `renbrowser.db`
- Display name and version labels

## Build and packaging

- `Taskfile.yml` : developer commands
- `build/` : per-OS packaging (Linux AppImage, Windows NSIS, macOS, Android, Docker)
- `build/config.yml` : Wails project config

## CI

GitHub Actions runs Go tests, frontend checks, security scans, desktop and server smoke builds, and release artifacts. See `.github/workflows/`.

## Extension points

1. **Plugins** : manifest-driven UI and schemes
2. **Themes** : JSON token files
3. **Community interfaces** : Reticulum config snippets in Settings

## Next steps

- [Development](development.md) to build locally
- [Extensions](extensions.md) for plugin API surface
- Source tree in the repo `README.md` layout table

# Getting started

Ren Browser lets you open NomadNet pages over Reticulum. Think of it as a small browser built for the mesh, not the public web.

## What Ren Browser does

- Opens `.mu` pages and other NomadNet content from mesh destinations
- Shows nodes you can reach through your Reticulum interfaces
- Keeps tabs, history, bookmarks, and settings on your machine (or in the browser when you use public server mode)
- Renders Micron markup with a WASM parser and built-in themes
- Supports extensions that add panels, URL schemes, and commands

## What you need first

Before Ren Browser can load a remote page, you need a working **Reticulum** setup on the same machine (or mounted into a Docker container for server mode).

Ren Browser reads Reticulum config from `~/.reticulum-go/` by default. You can point at another file with `--config` or the `REN_BROWSER_CONFIG` environment variable.

You do **not** need a traditional internet connection to browse NomadNet pages on the mesh. You do need at least one interface that can reach other Reticulum nodes.

## Desktop or server

| Mode | Best for |
|------|----------|
| **Desktop** (default) | Daily use on Linux, Windows, or macOS. Native window, local SQLite database. |
| **Server** | Homelab, Docker, or a machine that already runs Reticulum. You open Ren Browser in another browser at `http://host:8080`. |
| **Android** | Mobile builds when your release includes an APK. Same core browsing features in a touch layout. |

## First launch checklist

1. Install Reticulum and create or copy a config under `~/.reticulum-go/`
2. Install Ren Browser from [releases](https://github.com/Quad4-Software/Ren-Browser/releases) or build from source
3. Start Ren Browser and wait for Reticulum to connect on your interfaces
4. Open **Discovery** or type a 32-character destination hash in the address bar
5. Visit `about:` to confirm version, config path, and data directory

## Built-in pages (no mesh required)

These work even when you are offline from the mesh:

| Address | Purpose |
|---------|---------|
| `about:` | App version, build info, paths |
| `license:` | MIT license text |
| `editor:` | Built-in Micron editor |

Type `settings` in the address bar or press the settings shortcut to open preferences.

## Next steps

- [Installation](installation.md) if you have not installed yet
- [Reticulum setup](reticulum-setup.md) if pages fail to load or Discovery stays empty
- [Using the browser](using-the-browser.md) for everyday browsing

# Changelog

All notable changes to Ren Browser are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-07-08

First stable release.

### Added

- Wails v3 desktop shell with Svelte 5 frontend
- NomadNet page browsing over Reticulum (`quad4/reticulum-go` v0.9.8)
- Micron-first rendering via `micron-parser-go` with JS, Go, and WebAssembly parser options
- Tabbed browser chrome with discovery, history, downloads, devtools, settings, and plugins
- Unified search across history, discovered nodes, and favorites (`mod+shift+f`)
- Discovery favorites filter and per-node hop badges
- Temporary search-term highlighting when opening a page from search
- SQLite persistence for nodes, history, favorites, tabs, and page cache
- Multiple Reticulum identities: create, import, export, rename, and switch
- Micron editor with live preview
- Document reader for EPUB and PDF with in-document search and table of contents
- Extension system with JS and WASM plugins, permission grants, signing, and verification
- Dark/light themes with JSON import/export and custom accent tokens
- Localized UI: English, German, Spanish, and Russian
- Headless server mode, HTTP auth middleware hooks, and Docker deployment path
- Android APK builds with optional release signing and local APK sharing over HTTP
- Linux AppImage, `.deb`, `.rpm`, Flatpak, Windows (portable + NSIS), and macOS packaging
- Headless server binaries for Linux, FreeBSD, OpenBSD, NetBSD, and Windows (including legacy Windows 7/8 toolchain build)
- Landlock sandboxing on Linux with status reporting in Settings
- Per-interface Reticulum hot reload with tx/rx stats in Settings
- Custom frameless window controls (minimize, maximize, close) on desktop
- Mobile layout with bottom navigation, pull-to-refresh, edge back gesture, and tab sheet
- Micron layout preservation setting for fixed-width art and menus
- Community Reticulum interface import from bundled directory
- GitHub Actions CI: Go, frontend, security scans, reproducible server binary, and multi-platform release workflows
- GHCR publishing for the server image
- Nightly and beta release channels

[0.1.0]: https://github.com/Quad4-Software/Ren-Browser/releases/tag/v0.1.0

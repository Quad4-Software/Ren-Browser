# Changelog

All notable changes to Ren Browser are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-07-10

### Included

- Wails v3 desktop shell with Svelte 5 frontend
- NomadNet page browsing over Reticulum (`quad4/reticulum-go` v0.9.9)
- Micron-first rendering via `micron-parser-go` with JS, Go, and WebAssembly parser options
- Large micron pages prefer server Go HTML to avoid UI-thread re-parse lag
- Micron force-monospace for ASCII alignment, with preserve-layout as CSS-only horizontal scroll
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
- Optional overlay side panels (Appearance) so panels float over the page instead of shrinking it
- Localized UI: English, German, Spanish, Russian, Japanese, and Chinese
- Deep links via `renbrowser://` and `rns://` on desktop, Android, and iOS
- First-run setup for community interfaces and Reticulum config
- Settings toggles for shared instance and transport, with a mobile warning before enabling transport
- Shared-instance mode reporting (disabled, serving, or connected) in Settings
- Reset browser and restart Reticulum from Settings
- Built-in self-check diagnostics (`--self-check`)
- Path discovery with repeated nudges, dead-route expiry, and refresh between fetch retries
- Wake path invalidation after mobile suspend so reloads rediscover instead of reusing zombie routes
- Community interface seeding that prefers online clearnet TCP/backbone, dedupes endpoints, and picks 6 uplinks by default
- Live interop tests that seed community uplinks into an isolated config for CI
- Regression coverage for wake path invalidation, large micron Auto to Go, and ASCII force-monospace
- Headless server mode, HTTP auth middleware hooks, and Docker deployment path
- Android APK builds (including universal APK), optional release signing, and local APK sharing over HTTP
- Experimental iOS app packaging
- Linux AppImage, `.deb`, `.rpm`, Flatpak, Arch PKGBUILD, Nix flake, Windows (portable and NSIS), and macOS packaging
- Headless server binaries for Linux, FreeBSD, OpenBSD, NetBSD, and Windows (including legacy Windows 7/8 toolchain build)
- Landlock sandboxing on Linux with status reporting in Settings
- Per-interface Reticulum hot reload with tx/rx stats in Settings
- Custom frameless window controls (minimize, maximize, close) on desktop
- Mobile layout with bottom navigation, pull-to-refresh, edge back gesture, and tab sheet
- Micron layout preservation setting for fixed-width art and menus
- Community Reticulum interface import from bundled directory
- User docs in English, German, Spanish, Russian, Japanese, and Chinese
- GHCR publishing for the server image
- Nightly and beta release channels

[0.1.0]: https://github.com/Quad4-Software/Ren-Browser/releases/tag/v0.1.0

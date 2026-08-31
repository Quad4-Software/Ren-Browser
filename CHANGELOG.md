# Changelog

All notable changes to Ren Browser are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] [unreleased] - 2026-09-TBD

### Changed

- Bump reticulum-go to v1.1.0
- Refresh Go and frontend dependencies
- Stop tracking Android libwails.so and other build junk to shrink clones

## [0.3.0] [released] - 2026-08-21

### Fixed

- Keep mobile WebView from zooming or shifting chrome when the keyboard opens
- Stamp Android APK versionName and versionCode from the release version so Obtainium updates replace the installed app
- Restore split tab view on desktop. It was closed immediately by a mobile-layout effect
- Fix Android crash when minimizing and reopening the app caused by Go flag re-registration on activity recreation
- Run BrowserService cleanup on Wails shutdown by matching the ServiceShutdown interface
- Cap micron heading indent depth at 16 so nested section markers cannot emit unbounded CSS margins
- Reject Argon2id hashes with invalid cost parameters in VerifyPassword instead of panicking

### Changed

- Upgrade Wails from v3.0.0-alpha2.111 to v3.0.0-beta.8
- Desktop side panels can be resized by dragging the divider
- Speed up navigation by skipping unused panel refreshes and avoiding deep reactivity on tab page bodies
- Community interfaces load only from the build-time bundled snapshot
- Micron parser WASM is bundled at build time with no runtime GitHub download
- Keep tab page bodies on disk instead of SQLite, and drop inactive tab HTML from RAM
- Spill large page-cache entries to disk instead of keeping them in RAM
- Cap stored discovery nodes and browsing history so those tables cannot grow without bound

## [0.2.2] [released] - 2026-08-14

### Added

- Right-edge swipe goes forward in mesh history on mobile, matching the existing left-edge back gesture
- Haptic feedback on Android and iOS when a route is discovered, a download finishes, or the active identity is switched

### Security

- Improve HTML sanitization in page rendering

### Fixed

- Prevent bitrate calculation overflow in NomadNet by capping values at maximum integer limit
- Increase wait time to sixty seconds for frontend developer server connection to prevent startup timeouts
- Ensure safer file editing in vendor synchronization by using temporary files during version updates

### Changed

- Remove live community interface directory fetching
- Adjust NomadNet fetch budget calculations and connection timeout thresholds
- Update Go runtime and toolchain version to 1.26.6 in container, configuration, and workflows
- Upgrade web dependencies including Vite to 8.2.1, DOMPurify to 3.4.13, PostCSS to 8.5.26, and PDF.js to 6.2.108
- Standardize vendoring operations and update internal Go dependencies

## [0.2.1] [released] - 2026-07-25

### Security

- Harden HTML sanitization for document and micron content (XSS-oriented fixes and regression oracles)
- Safer download filenames and path handling to block traversal-style names
- Tighter micron rendering, navigation guards, tab preview srcdoc, and URL handling for untrusted mesh content

### Added

- Default Reticulum client limits on first config creation: in-memory paths, known destinations, soft memory budget, and packet hashlist cap (`internal/rns/config.go`)
- Exploratory security oracle tests for browser, micron, and document HTML surfaces

### Changed

- Bump `quad4/reticulum-go` to v1.0.1
- NomadNet page fetch uses `Link.RequestLimited` for max response size (matches reticulum-go 1.0.1)
- Updated `renbrowser.rsm` tree inventory for release signing

### Fixed

- Dependency `brace-expansion` pinned to 2.1.2 in workspace lockfiles

## [0.2.0] [released] - 2026-07-19

### Added

- Android back button/gesture uses mesh history instead of quitting the app
- Pull-to-refresh is more reliable on Android WebView
- Settings → Security shows Flatpak, WebKit bwrap, container, and Android status
- Tree inventory signing/verification (`renbrowser.rsm`) for release integrity checks
- `rngit-release --skip-build` when artifacts are already built
- Landlock ABI probing (IOCTL_DEV, IPC scopes, multithreaded TSYNC) and seccomp-bpf denylist hardening on Linux

### Fixed

- Micron WASM loader waits for parser readiness more cleanly
- Tab preview popover no longer hides under other chrome
- Docs call out creating host dirs for Reticulum config on install

### Changed

- Bump `quad4/reticulum-go` to v1.0.0
- README screenshot block removed so it reads better on the rngit nomadnet

## [0.1.0] [released] - 2026-07-10

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

[0.3.0]: https://github.com/Quad4-Software/Ren-Browser/releases/tag/v0.3.0
[0.2.2]: https://github.com/Quad4-Software/Ren-Browser/releases/tag/v0.2.2
[0.2.1]: https://github.com/Quad4-Software/Ren-Browser/releases/tag/v0.2.1
[0.2.0]: https://github.com/Quad4-Software/Ren-Browser/releases/tag/v0.2.0
[0.1.0]: https://github.com/Quad4-Software/Ren-Browser/releases/tag/v0.1.0

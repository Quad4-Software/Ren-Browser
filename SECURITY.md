# Security Policy

## Reporting a vulnerability

If you believe you have found a **security vulnerability** in Ren Browser, please report it privately so it can be fixed before wider disclosure.

**Preferred contact (in order):**

1. **LXMF**: `f489752fbef161c64d65e385a4e9fc74`

Include enough detail to reproduce or understand the issue (what version or build you used, what you expected, what happened). Do not open a public issue for unfixed vulnerabilities.

**Not security (legal, licensing, general questions):** see [`LEGAL.md`](LEGAL.md).

---

Ren Browser is meant to run on **systems and networks you trust** (for example your own desktop, a LAN, or a VPN you control).

The **headless server** (`renbrowser-server`) has **no built-in authentication**. If you expose it on the public internet, you accept much higher risk (automated scanning, abuse of Reticulum interfaces, and overload of a single-process app). If you must expose it: put it **behind a reverse proxy** with sensible access controls, use **HTTPS** with a valid certificate for the public name, and **restrict who can reach the port** (firewall, VPN, or proxy rules). Keep the application updated.

The **desktop app** uses the system webview (Wails v3). Treat installed plugins and downloaded files like any other local software: only install plugins from sources you trust.

### What you download should match what we built

Official release binaries and packages are built in **automation on GitHub** or on RNGit releases. Each tagged release is intended to ship:

- **Installable files** (Linux AppImage and binary, Windows executable, macOS app bundle, headless server binary, and Android APK when the pipeline produces it) from that tag.
- A **`SHA256SUMS.txt`** file listing checksums for release assets so you can verify downloads.
- **Software Bill of Materials** (`renbrowser-sbom.spdx.json` and `renbrowser-sbom.cyclonedx.json`) generated with Trivy (`task sbom`). Ad-hoc SBOM generation is available via `workflow_dispatch` on `.github/workflows/security.yml`.

**Docker images** for `renbrowser` published to GitHub Container Registry are built in CI with **build provenance and an SBOM** attached by Docker Buildx.


### Source tree integrity (`.rsm`)

The repository root includes a signed rnid message file, `renbrowser.rsm`. It embeds a SHA-256 inventory of git-tracked files except itself and paths under any `vendor/` tree. CI verifies the signature against the required signer identity `e46112d44649266d71fe2193e00a4710`, then re-hashes file bytes. Jobs also recheck the inventory at the end so a compromised runner cannot silently add or modify tracked files.

Verify locally:

```bash
make tree-rsm-verify
```

Maintainers regenerate the signature after intentional tree changes (requires a private identity file that hashes to the signer above, never commit `*.rid`):

```bash
export RNS_ID_PATH="$HOME/.local/share/reticulum-go/reticulum-go-release.rid"
make tree-rsm-sign
```

Enable the tracked pre-commit hook so commits that change inventory paths resign `renbrowser.rsm` automatically when that identity is available:

```bash
make hooks-install
```

Skip one commit with `SKIP_TREE_RSM_HOOK=1`.

### Practical tips

- Prefer **official download pages** or **GitHub Releases** for your copy of the app.
- For Docker, prefer an image referenced by **digest** (`@sha256:…`) once you trust a given build, not only by a moving tag.
- If something claims to be Ren Browser but does not match published checksums or verification steps, treat it as **untrusted**.

---

## For security professionals and auditors

### Product controls (high level)

- **Desktop (Wails v3):** The UI runs in the platform webview with Go bindings to application services. The app does not expose an arbitrary Node.js runtime to page content.
- **Server mode:** Optional reverse-proxy integration (`--trust-proxy`, `--base-path`) for deployment behind TLS terminators. No session or login layer is included; perimeter security is the operator's responsibility.
- **Data at rest:** Application state is stored in SQLite under `~/.renbrowser/`. Reticulum identity and transport data live under your Reticulum config directory.
- **Plugins:** Third-party plugins declare required **permissions** in their manifest; the host enforces the permission set at runtime. WASM plugins run in a constrained runtime with explicit capability grants.
- **External code (SRI):** Micron parser WebAssembly and its `wasm_exec.js` companion are verified with **SHA-384 Subresource Integrity (SRI)** before execution. If a hash mismatch is detected, the code is blocked and an error is reported. This reduces the risk of a tampered on-disk WASM binary being loaded even when upstream files are replaced locally.

### PDF and EPUB in-browser viewing

Ren Browser can open **PDF** and **EPUB** files from the **Downloads folder** inside the app (`document:` URLs). This is separate from page rendering: document bytes are read locally by the Go backend and passed to the UI as base64, they are **not** executed as page scripts.

**Backend (opening a file)**

- **Download directory only:** `document:` paths are resolved under the configured download folder. `validateDownloadPath` rejects paths outside that directory (including `..` traversal).
- **Format gate:** Only content detected as PDF or EPUB (by extension and/or magic bytes) is served for in-browser viewing.
- **Size cap:** Document reads are limited by `REN_BROWSER_MAX_PAGE_BYTES` (default **8 MiB**). Oversized files are rejected before loading into the viewer.
- **EPUB zip integrity:** Incomplete EPUB downloads (missing zip end-of-central-directory) may be repaired server-side so a finished download can be opened. Invalid or non-zip data is rejected.
- **No re-fetch for open tabs:** `document:` URLs are excluded from automatic mesh file download logic so a local document tab does not pull a second copy from the network on restart.

**PDF viewer (frontend)**

- Pages are rendered with **pdf.js** into a **`<canvas>`** inside a dedicated **sandboxed iframe** (`sandbox="allow-same-origin"` — **no** `allow-scripts`).
- The iframe document uses a strict **Content-Security-Policy** (`script-src 'none'`, `connect-src 'none'`, no frames/objects/forms).
- The PDF worker is loaded from the **bundled** pdf.js worker; loading options disable worker fetch, streaming, range requests, and system font face loading to avoid network I/O while parsing.
- Parse and render operations have **timeouts** to limit CPU/time spent on hostile files.

**EPUB viewer (frontend)**

- The EPUB zip is parsed in the UI; only **spine** chapters from the package manifest are shown.
- Zip-internal paths are normalized (`..` segments collapsed). **External** `http:`, `https:`, and other scheme URLs in images are stripped.
- Chapter HTML is sanitized with **DOMPurify** before display: scripts, iframes, objects, embeds, forms, media elements, and most external URLs are removed. Allowed image sources are **`blob:`** (in-archive assets) and **`data:image/...`** only.
- Links are deactivated (hrefs removed; pointer events disabled in the reader shell).
- Sanitized HTML is shown in the same **sandboxed iframe + strict CSP** model as above (`script-src 'none'`, no network).

**Residual risk and operator expectations**

- In-browser PDF/EPUB viewing is a **convenience layer**, not a full document sandbox comparable to a dedicated PDF reader or OS viewer. Parser and rendering bugs in pdf.js, JSZip, or the sanitization stack remain possible.
- **Only open documents from sources you trust.** Malicious PDFs or EPUBs may still attempt denial-of-service (very large or pathological files are partially mitigated by size limits and timeouts, not eliminated).
- For maximum isolation, open untrusted files with an external application instead of the built-in viewer.

### External network connections

Ren Browser does **not** include telemetry, analytics, crash reporting, or an in-app auto-update checker. The main browser UI **blocks** clearnet navigation (`http://`, `https://`, `ws://`, `wss://`) for page content.

The table below lists every **intentional** outbound connection path in application code. Third-party plugins and user-configured Reticulum peers can reach additional hosts beyond this list.

#### Runtime - automatic or app-initiated (clearnet HTTP/HTTPS)

| When | Destination | Purpose | Code |
|------|-------------|---------|------|
| Settings → Community interfaces refresh | `https://directory.rns.recipes/api/directory/submitted?search=&type=&status=online` | Fetch live Reticulum community interface directory. Falls back to an embedded snapshot on failure. | `internal/rns/community.go` |
| First Reticulum config creation | None (transport and sharing disabled by default) | Initialize empty Reticulum configuration. | `internal/rns/config.go` |
| Settings → Micron WASM Manager (user adds a GitHub release) | `https://github.com/Quad4-Software/Micron-Parser-Go/releases/download/{tag}/micron-parser-go.wasm` and `.../SHASUMS256.txt` | Download and SHA-256–verify an optional Micron parser WASM binary. | `internal/micronwasm/fetch.go` |

The community directory URL is also fetched at **build time** to refresh the embedded snapshot (`build/scripts/fetch-community-directory.mjs`). Override for builds only: `REN_BROWSER_COMMUNITY_DIRECTORY_URL`.

#### Runtime - user-initiated or permission-gated

| When | Destination | Purpose | Code |
|------|-------------|---------|------|
| Plugin with `network.fetch` permission | Any `http://` or `https://` URL the plugin requests | Plugin HTTP client (`PluginFetch` / `ctx.network.fetch`). Manifest `network.endpoints` is disclosure for install review, not a runtime allowlist. | `internal/plugins/wasm_http.go`, `internal/app/plugin_fetch.go` |
| Micron Translator extension (bundled; requires `network.fetch`) | `https://translate.googleapis.com/translate_a/single?...` | Google Translate backend (default). | `extensions/micron-translator/wasm/main.go` |
| Micron Translator extension | `https://libretranslate.com/translate` (or user-configured LibreTranslate base URL) | LibreTranslate backend (optional). | `extensions/micron-translator/wasm/main.go`, `extensions/micron-translator/settings.js` |
| Android APK share | `http://{local-wifi-ip}:{port}/{filename}` | Temporary LAN HTTP server so another device on the same network can download a shared APK. | `internal/app/share_apk_android.go` |
| Android `OpenURL` | User-chosen URI | Opens the system browser or handler for a link (destination depends on context). | `internal/app/downloads_android.go` |

#### Runtime - mesh / Reticulum (not clearnet HTTP)

| When | Destination | Purpose | Code |
|------|-------------|---------|------|
| NomadNet browsing, file fetch, LXMF | Reticulum destination hashes over configured transports (TCP, I2P, Yggdrasil, WebSocket, etc.) | Load `.mu` pages and mesh files. No hardcoded internet URLs; peers come from Reticulum config. | `internal/nomadnet/`, `internal/app/browser_service.go` |
| User or admin Reticulum config | Arbitrary `target_host`, `remote`, WebSocket URLs, I2P peers, etc. | Transport interfaces defined in `~/.reticulum-go/` or `REN_BROWSER_CONFIG`. | Reticulum config (operator-controlled) |

### Build, supply chain, and transparency

- **CI:** Automated pipelines on GitHub Actions run Go and frontend tests, **gosec** static analysis, **Trivy** filesystem and Dockerfile configuration scans, brand consistency checks, and server/desktop build smoke tests. **CodeQL** analysis runs on a separate schedule/workflow. Third-party GitHub Actions are referenced with **pinned commit SHAs** (documented beside each workflow) to reduce unexpected upgrades. SPDX and CycloneDX SBOMs are produced with Trivy (`task sbom` / `make sbom`) and attached to GitHub Releases from the desktop build workflow; ad-hoc generation uses `.github/workflows/security.yml`.
- **Releases:** Tagged release artifacts for Linux, Windows, macOS, and the headless server are produced in CI and published with **SHA256** checksums and SBOM files. Android release APKs are built when the Android pipeline is enabled for the tag.
- **Containers:** Server images are built with provenance and SBOM generation enabled in the Docker workflow.

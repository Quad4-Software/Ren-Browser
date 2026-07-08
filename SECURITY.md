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

Official release binaries and packages are built in **automation on GitHub**, not by hand. Each tagged release is intended to ship:

- **Installable files** (Linux AppImage and binary, Windows executable, macOS app bundle, headless server binary, and Android APK when the pipeline produces it) from that tag.
- A **`SHA256SUMS.txt`** file listing checksums for release assets so you can verify downloads.

**Docker images** for `renbrowser` published to GitHub Container Registry are built in CI with **build provenance and an SBOM** attached by Docker Buildx.

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

### Build, supply chain, and transparency

- **CI:** Automated pipelines on GitHub Actions run Go and frontend tests, **gosec** static analysis, **Trivy** filesystem and Dockerfile configuration scans, brand consistency checks, and server/desktop build smoke tests. **CodeQL** analysis runs on a separate schedule/workflow. Third-party GitHub Actions are referenced with **pinned commit SHAs** (documented beside each workflow) to reduce unexpected upgrades.
- **Releases:** Tagged release artifacts for Linux, Windows, macOS, and the headless server are produced in CI and published with **SHA256** checksums. Android release APKs are built when the Android pipeline is enabled for the tag.
- **Containers:** Server images are built with provenance and SBOM generation enabled in the Docker workflow.

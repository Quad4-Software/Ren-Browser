# Security

Ren Browser is meant for systems and networks you trust. This page summarizes safe use, download verification, and how to report problems.

## Trust boundaries

| Surface | Risk |
|---------|------|
| Desktop app | Local webview with Go bindings. No Node.js in page content. |
| Server mode | Open HTTP port. No login included. |
| Plugins | Run with declared permissions. WASM plugins are sandboxed with caps. |
| Mesh content | Untrusted like any network content. Micron HTML is sanitized unless a plugin requests `render.unsanitized`. |

## Server mode

`renbrowser-server` has **no authentication**. If you expose it on the internet you risk:

- Automated scanning
- Abuse of your Reticulum interfaces
- Overload of a single-process app

If you must expose it:

1. Put it behind a reverse proxy with access controls
2. Use HTTPS with a valid certificate
3. Restrict the port with firewall or VPN rules
4. Keep the app updated

See [Server mode](server-mode.md) for proxy flags.

## Desktop plugins

Only install extensions from people or projects you trust. Read the permission list in **Settings → Extensions** before enabling.

## Verify downloads

Official builds come from [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases) and GitHub Actions CI.

Each release should include `SHA256SUMS.txt`. Check your file:

```sh
sha256sum -c SHA256SUMS.txt
```

For Docker, prefer pinning by digest (`@sha256:...`) after you trust a build. Images on GHCR include build provenance and an SBOM from Docker Buildx.

If a binary does not match published checksums, treat it as untrusted.

## Subresource integrity for WASM

Micron parser WebAssembly and its `wasm_exec.js` companion are checked with SHA-384 SRI before execution. A hash mismatch blocks the code and shows an error.

## Data at rest

- Application state: SQLite under `~/.renbrowser/`
- Reticulum keys: your Reticulum config directory
- Server public mode: some data only in browser `localStorage` on each client

Encrypt disks at the OS level if the machine is shared or portable.

## Reporting vulnerabilities

Do **not** open a public GitHub issue for unfixed security bugs.

**Preferred contact:**

1. LXMF: `f489752fbef161c64d65e385a4e9fc74`

Include version, platform, steps to reproduce, and impact.

Legal and licensing questions go to [LEGAL.md](../../LEGAL.md) (`legal@quad4.io`), not the security channel.

## CI and supply chain (overview)

GitHub Actions runs tests, gosec, Trivy scans, and CodeQL on a schedule. Third-party Actions are pinned to commit SHAs in workflows.

## Next steps

- [Extensions](extensions.md) permission list
- [Server mode](server-mode.md) deployment
- Repo root [SECURITY.md](../../SECURITY.md) for the canonical policy

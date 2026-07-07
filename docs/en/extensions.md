# Extensions

Ren Browser supports plugins that add URL schemes, sidebar panels, commands, themes, settings pages, and devtools tabs.

## Installing extensions

### From Settings

1. Open **Settings → Extensions**
2. Choose **Install extension**, then pick a **.zip**, **folder**, or **bundled .wasm module**
3. Review the install preview:
   - Requested permissions (you can disable individual permissions before install)
   - External URLs the extension may contact (scanned from the manifest and package files)
   - Publisher signature status (unsigned, signed, signed by a trusted publisher, or invalid)
   - Security assessment warnings
   - Bundled UI languages (when the extension ships `locales/*.json`)
4. Confirm and enable the extension

Extensions that request `network.fetch` show a confirmation dialog listing detected endpoints. Endpoints remain visible even if you disable `network.fetch` during install so you can see what the package would contact when that permission is granted.

### Manual install

Unpack a plugin into:

```
~/.renbrowser/plugins/<id>/
```

The folder must contain `renbrowser.plugin.json`. The `id` in the manifest should match the folder name.

## Example extensions

The repo includes `extensions/hello-extension/`:

- Registers the `hello:` URL scheme
- Adds a **Hello** sidebar panel
- Defines a **Say hello** command with `mod+shift+h`

Use it as a template when you write your own plugin.

`extensions/micron-translator/` translates Micron (`.mu`) pages using Google Translate (public endpoint) or a LibreTranslate instance (URL and optional API key in the sidebar panel). Commands: **Translate Micron page** (`mod+shift+t`) and **Restore original** (`mod+shift+r`).

## Manifest file

File name: `renbrowser.plugin.json`

Required fields:

| Field | Purpose |
|-------|---------|
| `manifestVersion` | Currently `1` |
| `id` | Unique id (`a-z`, `A-Z`, `0-9`, `.`, `-`, 3 to 128 chars) |
| `name` | Display name |
| `version` | Semver string |
| `main` | Frontend entry script (optional if only backend) |
| `permissions` | Capability list (see below) |

Optional fields include `description`, `author`, `license`, `engines`, `backend`, `network`, and `contributes`.

### Engine constraint

```json
"engines": { "renbrowser": ">=0.1.0" }
```

The host refuses to load the plugin if your app version is too old.

### Network endpoints

Extensions that use `network.fetch` should declare contacted hosts or URLs:

```json
"network": {
  "endpoints": [
    "https://api.example.com/",
    "User-configured service URL"
  ]
}
```

At install time RenBrowser also scans `.js`, `.go`, `.wasm`, and other package files for `http`/`https` URLs and lists anything it finds alongside manifest entries.

### Contributions

| Type | Purpose |
|------|---------|
| `urlSchemes` | Handle custom schemes |
| `panels` | Sidebar or other panel slots |
| `commands` | Command palette entries and keybinds |
| `themes` | Extra theme JSON files |
| `settings` | Settings sub-pages |
| `devtools` | DevTools tabs |
| `renderers` | Custom renderers for MIME types or extensions |

## Permissions

Plugins must declare what they need. Known permissions:

| Permission | Allows |
|------------|--------|
| `storage.plugin` | Private key-value storage for the plugin |
| `navigation.read` | Read current URL and tab info |
| `navigation.write` | Trigger navigation |
| `network.fetch` | Fetch over allowed network APIs |
| `events.emit` | Emit host events |
| `events.subscribe` | Listen to host events |
| `devtools.network` | Extra network detail in DevTools |
| `render.unsanitized` | Skip some HTML sanitization (dangerous) |

The host enforces permissions at runtime. Permissions you disable at install are stored per extension and are not granted to JS `ctx.network.fetch` or WASM `http_fetch`.

## Publisher signatures

Extensions may ship an Ed25519 signature in `renbrowser.plugin.rsg` (compatible with Reticulum `rnid` tooling). Signed packages with an invalid signature cannot be installed.

The install preview and extension list show badges:

| Badge | Meaning |
|-------|---------|
| Unsigned | No signature file present |
| Signed | Valid signature from a Reticulum identity |
| Trusted | Signed by a publisher in the trusted list |
| Tampered | Extension files changed outside RenBrowser (extension is disabled until you re-enable it) |

During install you can choose **Trust this publisher identity** to add a valid signer to your user trusted list (`~/.renbrowser/trusted_publishers.json`). RenBrowser also ships a small bundled trusted list. The user list is protected by a digest stored in the profile database; external edits without updating the database are detected.

Sign a directory or zip with `build/scripts/sign-extension.sh` (requires Python `rnid`).

## Plugin UI translations

Extensions may bundle their own UI strings under `locales/<code>.json` (for example `locales/en.json`). Panel titles and commands can use `%key.path%` placeholders in the manifest; the host loads catalogs from `/_plugins/<id>/locales/<code>.json`.

The install preview lists bundled locale codes when present.

## Frontend entry script

A typical `main.js` exports:

- `activate(ctx)` : subscribe to events, register UI
- `deactivate()` : cleanup
- `mount(el)` : render sidebar panel HTML
- `handleScheme(url)` : for URL scheme handlers

Plugins with `network.fetch` may call `ctx.network.fetch()` for HTTP GET/POST to public `http`/`https` URLs when that permission was granted at install. Check `ctx.capabilities.networkFetch` before starting network-backed work.

Plugins with a `backend` WASM module may call `ctx.wasm.call(export, input)` to run exported functions such as `translate_micron`. Use `ctx.content.getActivePage()`, `ctx.content.renderRaw(path, raw)`, and `ctx.content.updateActivePage()` to re-render the active tab after transforming Micron source.

Use `ctx.i18n.t("key")` for strings from the extension locale files.

## Bundled WASM modules

A distributable extension can be shipped as one `.wasm` file. The module carries custom sections:

- `renbrowser.plugin` — manifest JSON (`renbrowser.plugin.json`)
- `renbrowser.files` — map of relative paths to UTF-8 file contents (for example `main.js`, `locales/en.json`)
- `renbrowser.signature` — optional RSG signature bytes

Install from **Settings → Extensions → Install extension → Choose .wasm module**. The host unpacks metadata into the plugins directory and keeps the WASM binary as the manifest `backend`.

`extensions/micron-translator/` ships `translator.wasm` (TinyGo). Rebuild with `extensions/micron-translator/build-wasm.sh`, or bundle with `go run ./extensions/micron-translator/bundle` after building.

## WASM backend

Plugins may set `backend` to a WASM module path for heavier logic. WASM plugins run in a constrained runtime with explicit grants.

The host provides a `renhost` module with `http_fetch` when `network.fetch` was granted at install. Exported functions such as `translate_micron(in_ptr, in_len) -> out_len` read JSON input and write JSON output in linear memory.

Safeguards include per-call network request limits, WASM call timeouts, and input size caps. Network-heavy exports are blocked entirely when `network.fetch` is not granted.

## DevTools

When **Developer tools → Network** is open, outbound HTTP requests made by extensions (JS `PluginFetch` and WASM `http_fetch`) appear in the log with source **Extension fetch**, status code, and duration.

## Integrity and tampering

After install RenBrowser stores a cryptographic hash of each extension's file payload (excluding the signature file). If files change on disk outside the app, the extension is disabled and marked **Tampered**. Re-enabling accepts the current files and refreshes the stored hash.

## Security notes

- Only install plugins from sources you trust
- Read permissions and detected network endpoints before confirming install
- Prefer signed extensions from publishers you recognize
- Treat plugins like any local program with access to your profile data

## Next steps

- Source reference: `internal/plugins/manifest.go` in the repo
- [Security](security.md) for plugin threat model and signing
- [Development](development.md) to hack on the plugin host

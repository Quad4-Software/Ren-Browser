# Extensions

Ren Browser supports plugins that add URL schemes, sidebar panels, commands, themes, settings pages, and devtools tabs.

## Installing extensions

### From Settings

1. Open **Settings → Extensions**
2. Choose **Install extension**, then pick a **.zip**, **folder**, or **bundled .wasm module**
3. Confirm the manifest loads and permissions look correct
4. Enable the extension

### Manual install

Unpack a plugin into:

```
~/.renbrowser/plugins/<id>/
```

The folder must contain `renbrowser.plugin.json`. The `id` in the manifest should match the folder name.

## Example extension

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

Optional fields include `description`, `author`, `license`, `engines`, `backend`, and `contributes`.

### Engine constraint

```json
"engines": { "renbrowser": ">=0.1.0" }
```

The host refuses to load the plugin if your app version is too old.

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

The host enforces permissions at runtime. A plugin cannot use a capability it did not declare.

## Frontend entry script

A typical `main.js` exports:

- `activate(ctx)` : subscribe to events, register UI
- `deactivate()` : cleanup
- `mount(el)` : render sidebar panel HTML
- `handleScheme(url)` : for URL scheme handlers
- `mount(el)` : render sidebar panel HTML (called by the panel host when present)

Plugins with `network.fetch` may call `ctx.network.fetch()` for HTTP GET/POST to public `http`/`https` URLs (enforced by the host). Plugins with a `backend` WASM module may call `ctx.wasm.call(export, input)` to run exported functions such as `translate_micron`. Use `ctx.content.getActivePage()`, `ctx.content.renderRaw(path, raw)`, and `ctx.content.updateActivePage()` to re-render the active tab after transforming Micron source.

`extensions/micron-translator/` ships a TinyGo WASM backend (`translator.wasm`) that preserves Micron markup while translating visible text through Google Translate or LibreTranslate. Rebuild locally with `extensions/micron-translator/build-wasm.sh` when TinyGo is installed, or download prebuilt assets from [GitHub Releases](https://github.com/Quad4-Software/Ren-Browser/releases) (`renbrowser-micron-translator.wasm` for one-file install).

### Bundled WASM modules

A distributable extension can be shipped as one `.wasm` file. The module carries two custom sections:

- `renbrowser.plugin` — manifest JSON (`renbrowser.plugin.json`)
- `renbrowser.files` — map of relative paths to UTF-8 file contents (for example `main.js`)

Install it from **Settings → Extensions → Install extension → Choose .wasm module**. The host unpacks the metadata into the plugins directory and keeps the WASM binary as the manifest `backend`.

Bundle an extension with `go run ./extensions/micron-translator/bundle` after building `translator.wasm`.

The hello extension shows minimal versions of each.

## WASM backend

Plugins may set `backend` to a WASM module path for heavier logic. WASM plugins run in a constrained runtime with explicit grants.

The host provides a `renhost` module with `http_fetch` when `network.fetch` is declared. Exported functions such as `translate_micron(in_ptr, in_len) -> out_len` read JSON input and write JSON output in linear memory at the input pointer. TinyGo modules must export `_initialize`; the host calls it after instantiation.

Call from the frontend with `ctx.wasm.call("translate_micron", { body, settings })`.

## Security notes

- Only install plugins from sources you trust
- Read the permission list before enabling
- Treat plugins like any local program with access to your profile data

## Next steps

- Source reference: `internal/plugins/manifest.go` in the repo
- [Security](security.md) for plugin threat model
- [Development](development.md) to hack on the plugin host

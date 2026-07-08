# Navigation and URLs

Ren Browser accepts several URL shapes in the address bar. This page lists them and explains how normalization works.

## NomadNet destinations

A full NomadNet URL looks like:

```
<32-hex-chars>:/page/path.mu
```

Example:

```
a1b2c3d4e5f6789012345678abcdef01:/page/index.mu
```

The hash is a Reticulum destination identity in hexadecimal. The path after `:/` is the file on that NomadNet node.

## Shorthand hash

If you enter only a 32-character hex string, Ren Browser expands it to:

```
<hash>:/page/index.mu
```

This matches the usual NomadNet home page path.

## Paths without a hash

If the current context already has a destination, a path starting with `/page/` may resolve relative to that node. For a cold navigation, prefer the full `hash:/page/...` form.

## Deep links

Ren Browser registers the `renbrowser://` and `rns://` URL schemes so the OS can open the app from a link.

Examples:

```
renbrowser:about
renbrowser://about
renbrowser://open?url=about%3A
renbrowser://abb3ebcd03cb2388a838e70c001291f9/page/index.mu
renbrowser://rns/abb3ebcd03cb2388a838e70c001291f9/page/home.mu
rns://abb3ebcd03cb2388a838e70c001291f9/page/home.mu
```

Supported on Android, iOS, macOS, Windows, and Linux (desktop file / AppImage / packages). External `http://` and `https://` links are not accepted as deeplink targets.

## Built-in schemes

These schemes are handled inside the app. They do not use the mesh.

| Scheme | Aliases | Description |
|--------|---------|-------------|
| `about:` | `about` | Version, build, Reticulum config path, data directory |
| `license:` | `license` | Project license (MIT) |
| `editor:` | `editor` | Micron source editor |

Matching is case-insensitive. Trailing spaces are trimmed.

## Extension URL schemes

Installed extensions can register custom schemes in `renbrowser.plugin.json`. For example the hello extension registers `hello:`. See [Extensions](extensions.md).

## Settings and internal routes

The UI may use internal routes for panels. The address bar focuses on mesh and built-in schemes. Open **Settings** with the sidebar button or `Ctrl+,` / `Cmd+,`.

## Tab titles

Tab titles come from:

1. Page metadata when the node provides a title
2. Discovery display names when the hash matches a known node
3. A shortened hash or path as fallback

## History entries

Each navigation that loads content can create a history row with URL, title, destination hash, and timestamp. Built-in pages like `about:` are included.

## Link resolution on Micron pages

When you click a link on a rendered page:

- `about:`, `license:`, and `editor:` open locally
- Absolute mesh URLs navigate directly
- Relative paths combine with the current page destination

## Normalization rules (summary)

| You type | Normalized URL |
|----------|----------------|
| `about` | `about:` |
| `license` | `license:` |
| `editor` | `editor:` |
| `abcdef...` (32 hex) | `abcdef...:/page/index.mu` |
| Full `hash:/page/foo.mu` | unchanged |

## Next steps

- [Using the browser](using-the-browser.md) for tabs and panels
- [Discovery](discovery.md) to pick nodes without typing hashes

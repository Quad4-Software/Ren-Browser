# Troubleshooting

Common problems and what to try first.

## Reticulum will not start

**Symptoms:** Log line `reticulum start: ...`, Discovery empty, all mesh pages fail.

**Checks:**

1. Config path in `about:` matches where your files live
2. `REN_BROWSER_CONFIG` or `--config` points at a valid file
3. Interface definitions are syntactically correct
4. Keys and storage paths are readable by the user running Ren Browser

Fix Reticulum config outside the app, then reload from Settings or restart.

## Discovery is empty

See [Discovery](discovery.md) and [Reticulum setup](reticulum-setup.md).

Short list:

- Wait after connect for announcements
- Confirm peers exist on your interfaces
- Check firewall rules for UDP or TCP ports you configured

## Page load timeout

1. Verify the hash in the address bar
2. Open another node from Discovery
3. Confirm Reticulum shows traffic on interfaces in Settings
4. Retry after radio or path changes on mesh networks

## Database corrupt or will not open

**Symptoms:** Error about profile data, offer to reset database.

**Options:**

1. Restore `renbrowser.db` from backup ([Data and profiles](data-and-profiles.md))
2. Reset through the UI (destroys local tabs, history, favorites, settings)
3. Rename the bad file and let Ren Browser create a fresh database

Reticulum identity is unaffected by a browser DB reset.

## WASM or Micron parser error

If SRI check fails for Micron WASM:

1. Do not disable the check
2. Reinstall from official releases
3. If you built from source, run `task build` again without hand-editing `frontend/dist/vendor/`

## Server mode: blank page or wrong assets

1. Check `--base-path` matches your reverse proxy mount
2. Enable `--trust-proxy` when TLS terminates upstream
3. Confirm port mapping in Docker (`-p 8080:8080`)

## Server mode: shared history when you did not want it

Start with `--public-mode` so each browser keeps its own `localStorage` copy.

## Extension will not load

1. Manifest must be valid JSON in `renbrowser.plugin.json`
2. `id` must match folder name under `plugins/`
3. `engines.renbrowser` must be satisfied by your app version
4. Unknown permission strings cause load failure

Check Settings for the error string.

## Android build fails

1. Set `ANDROID_HOME`
2. Run `task android:install:deps`
3. Use API 34 and NDK r26+ as documented in [Installation](installation.md)

## Development: `task check` fails

| Area | Command |
|------|---------|
| Go format | `task fmt:go` |
| Go tests | `task test:go` |
| Frontend | `task frontend:check` |
| Security scan | `task gosec` |
| SBOM | `task sbom` |

Run `task check` before you send patches.

## Still stuck

1. Note your version from `about:`
2. Capture logs from terminal or Docker
3. Ask on your mesh community or send a detailed bug report through project channels

See [Contributing](contributing.md) for patch submission over LXMF.

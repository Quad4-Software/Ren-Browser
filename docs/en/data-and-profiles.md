# Data and profiles

Ren Browser stores bookmarks, history, tabs, settings, and discovery cache on disk. This page explains where that data lives and how profiles work.

## Default locations

| Item | Path |
|------|------|
| Data directory | `~/.renbrowser/` |
| Main database | `~/.renbrowser/renbrowser.db` |
| Plugins | `~/.renbrowser/plugins/<id>/` |
| Named profiles | `~/.renbrowser/profiles/<name>/renbrowser.db` |
| Legacy state (migrated) | `~/.renbrowser/state.json` |

On first launch after an upgrade, `state.json` imports into SQLite automatically.

## What SQLite holds

Typical tables and blobs include:

- Open tabs and session restore state
- Browsing history with timestamps
- Favorites
- Browser preferences and keybinds
- Cached discovery entries
- Theme selection and custom theme data

Corruption is detected on open. The UI may offer to reset the database. Reset removes local tabs, history, favorites, and settings.

## Reticulum data is separate

Identity keys and interface config live in your Reticulum directory (`~/.reticulum-go/` by default). Ren Browser reads that path but does not move your Reticulum identity into `~/.renbrowser/`.

## Named profiles

Start with `--profile NAME` or `REN_BROWSER_PROFILE=NAME` to use:

```
~/.renbrowser/profiles/NAME/renbrowser.db
```

Use profiles when you want separate histories on one account (work vs personal, or testing).

## Import and export

At startup only:

- `--export-profile /path/to/backup.json` writes profile data and exits
- `--import-profile /path/to/backup.json` merges or replaces from file

Environment mirrors: `REN_BROWSER_EXPORT_PROFILE`, `REN_BROWSER_IMPORT_PROFILE`.

Use export before major upgrades or when moving to a new machine.

## Server mode storage

| Mode | Tabs, history, favorites |
|------|--------------------------|
| Default server | Server-side SQLite in the server's `~/.renbrowser/` |
| `--public-mode` | Each client's browser `localStorage` |

Pick public mode when many users share one server instance.

## Theme import and export

Themes can be exported as JSON from Settings and imported on another install. Theme files are not the full profile, only appearance tokens.

## Plugins data

Extensions with `storage.plugin` permission get isolated storage keyed by plugin id. Uninstalling a plugin does not always remove its folder. Delete `~/.renbrowser/plugins/<id>/` manually if you want a clean removal.

## Android

Mobile builds use the same logical layout under the app sandbox. Paths differ by OS rules but the database schema matches desktop.

## Backup checklist

1. Stop Ren Browser
2. Copy `~/.renbrowser/renbrowser.db` (or your profile path)
3. Copy `~/.reticulum-go/` if you also want mesh identity
4. Copy `~/.renbrowser/plugins/` if you use extensions

Restore by placing files back before the next launch.

## Next steps

- [Settings](settings.md) for UI paths
- [Server mode](server-mode.md) for public mode
- [Troubleshooting](troubleshooting.md) if the database fails to open

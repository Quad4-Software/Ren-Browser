# Settings

Settings control how Ren Browser looks, connects to Reticulum, and behaves on your machine.

Open Settings from the sidebar or with `Ctrl+,` / `Cmd+,` (customizable in keybind settings).

## Appearance

### Themes

Ren Browser ships with dark and light themes. You can:

- Switch theme in Settings
- Import a theme JSON file
- Export your current theme for backup or sharing

Themes affect chrome colors, typography tokens, and Micron page styling where applicable.

### Window chrome

On desktop you can choose between a native title bar and a frameless window with custom minimize, maximize, and close controls. Frameless mode uses draggable regions at the top of the window.

## Reticulum

The Reticulum section shows:

- Active interfaces and their status
- Transmit and receive byte counters
- Config editor with hot reload

After you edit config, apply changes from Settings instead of restarting the app when possible.

## Keybinds

Each action can have a chord such as `mod+t` for new tab. `mod` means Ctrl on Windows and Linux, Cmd on macOS.

See [Keyboard shortcuts](keyboard-shortcuts.md) for defaults and how to record new bindings.

## Extensions

Manage plugins from **Settings → Extensions**:

- Install from a zip archive
- Install from a folder
- Enable or disable installed extensions
- View permissions each extension requested

See [Extensions](extensions.md) for manifest format and install paths.

## Profile and data

Settings surfaces paths for:

- SQLite database location
- Reticulum config path
- Plugin directory under `~/.renbrowser/plugins/`

For named profiles and import or export, see [Data and profiles](data-and-profiles.md).

## Browser preferences

Additional toggles may include:

- Native title bar vs frameless window
- Default panel behavior
- Options that sync with stored browser prefs in SQLite

Exact toggles can vary by version. When in doubt, check the label in the UI and this doc for your release tag.

## Mobile layout

On Android, Settings uses the same data but may group items for touch navigation. Core Reticulum and theme options remain available.

## Next steps

- [Keyboard shortcuts](keyboard-shortcuts.md)
- [Extensions](extensions.md)
- [Data and profiles](data-and-profiles.md)

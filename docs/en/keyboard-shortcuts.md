# Keyboard shortcuts

Default keybindings on desktop. You can change them in **Settings → Keybinds**.

`mod` means **Ctrl** on Windows and Linux, **Cmd** on macOS.

## Default bindings

| Action | Chord | Description |
|--------|-------|-------------|
| Focus address bar | `mod+l` | Move keyboard focus to the URL field |
| Reload page | `mod+r` | Reload the active tab |
| Developer tools | `mod+shift+i` | Open or close DevTools |
| Find in page | `mod+f` | Open find bar on the current page |
| Discovery panel | `mod+shift+d` | Open Discovery sidebar |
| Settings panel | `mod+,` | Open Settings |
| New tab | `mod+t` | Open a blank tab |
| New window | `mod+shift+n` | Open another window (when supported) |
| Close tab | `mod+w` | Close the active tab |
| Fullscreen | `f11` | Toggle fullscreen |

## Recording a new binding

1. Open **Settings → Keybinds**
2. Click the action you want to change
3. Press the new key combination
4. Conflicts with another action show in the UI. Resolve before saving.

While recording, other shortcuts are paused so the recorder sees only your new chord.

## Chord syntax

Bindings are stored as lowercase chords joined by `+`:

- `mod` : Ctrl or Cmd
- `shift` : Shift
- `alt` : Alt
- Final segment is the key name (`l`, `r`, `,`, `f11`, etc.)

Example: `mod+shift+d` is Ctrl+Shift+D on Linux.

## Extension keybinds

Extensions can contribute commands with optional `keybind` fields in `renbrowser.plugin.json`. Enabled extensions merge their commands into the host command palette and shortcut handling.

## Android

Hardware keyboards on Android follow the same chord rules when connected. Touch UI does not require shortcuts.

## Next steps

- [Settings](settings.md) for other preferences
- [Using the browser](using-the-browser.md) for panel overview

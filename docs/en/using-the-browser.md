# Using the browser

This page covers everyday use: tabs, the address bar, history, favorites, and page viewing.

## Window layout

The main areas are:

- **Tab bar** at the top. Drag tabs, pin them, or open split views.
- **Address bar** for destinations and built-in schemes.
- **Content area** where Micron pages render.
- **Side panels** for Discovery, History, Settings, DevTools, and extension panels.

On smaller screens a mobile navigation bar replaces some of the desktop chrome.

## Tabs

- **New tab**: default shortcut `Ctrl+T` / `Cmd+T` (see [Keyboard shortcuts](keyboard-shortcuts.md))
- **Close tab**: `Ctrl+W` / `Cmd+W`
- **Restore session**: open tabs are saved in the local database and restored on next launch

Pinned tabs stay at the front of the tab strip. Split view lets you show two tabs side by side.

## Address bar

Type a NomadNet destination or use a built-in scheme:

| Input | Result |
|-------|--------|
| 32-char hex hash | Opens `hash:/page/index.mu` |
| Full NomadNet URL | Opens as typed |
| `about:` | About page with version and paths |
| `license:` | License text |
| `editor:` | Micron editor |
| `settings` | Opens Settings (via UI routing) |

Press Enter to navigate. The bar also accepts focus via `Ctrl+L` / `Cmd+L`.

## Following links

Micron links on a page resolve relative to the current destination. Internal `about:`, `license:`, and `editor:` links work like normal navigation.

External mesh links use Reticulum destination syntax. If a link fails, check that the target node is reachable.

## History

Open the History panel from the sidebar or its keybind. You can:

- Search by title or URL
- See entries grouped by date
- Open a past page in the current or a new tab

History is stored locally in SQLite (desktop) unless you use server public mode.

## Favorites

Save nodes you visit often from the page context menu or Discovery. Favorites sync with the same store as history and tabs.

## Find in page

Press `Ctrl+F` / `Cmd+F` to search text on the current Micron page. Matches highlight in the content viewer.

## Developer tools

Press `Ctrl+Shift+I` / `Cmd+Shift+I` to open DevTools. Useful for:

- Inspecting render timing
- Viewing raw page source
- Debugging extension panels that contribute devtools entries

## Downloads

When a page or action offers a file, items appear in the downloads menu. Paths follow your OS download conventions on desktop.

## Page errors

If a page cannot load, you see an error state with a short message. Common causes:

- Destination unreachable
- Invalid Micron content
- Reticulum not connected

Use **Reload** (`Ctrl+R` / `Cmd+R`) after you fix connectivity.

## Micron editor

Open `editor:` to compose Micron in a built-in editor. Use this for local drafts before you publish to a NomadNet node.

## Next steps

- [Navigation and URLs](navigation-and-urls.md) for URL rules in detail
- [Discovery](discovery.md) to browse the mesh
- [Settings](settings.md) for themes and interfaces

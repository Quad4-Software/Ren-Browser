# Discovery

Discovery shows NomadNet nodes announced on your Reticulum interfaces. Use it when you do not already know a destination hash.

## Opening Discovery

- Click **Discovery** in the sidebar
- Use the keyboard shortcut (`Ctrl+Shift+D` / `Cmd+Shift+D` by default)

The panel lists nodes with display names, hashes, and metadata when the announcement includes it.

## What appears in the list

A node shows up when:

1. Reticulum receives an announcement for that destination
2. The announcement matches NomadNet or compatible node types the browser understands
3. Your interfaces can reach the path that carried the announcement

Discovery is live. The list updates as new announcements arrive and old ones expire.

## Opening a node

Click a row to navigate to that node's default page (usually `index.mu`). You can also copy the hash for use in the address bar.

## Favorites from Discovery

Many rows offer a way to add the node to favorites. Favorites are stored in your local profile database.

## Empty list

If Discovery stays empty, work through [Reticulum setup](reticulum-setup.md):

- Interfaces must be up in Settings
- You need connectivity to peers that carry announcements
- New joins can take a short time to populate

## Announcements vs reachability

Seeing a node in Discovery does not guarantee every page load will succeed. You still need a route to the destination for LXMF page requests over Reticulum.

If Discovery shows a node but pages fail, check transport paths and whether the remote node is online.

## Community interfaces

Settings may list community or shared interface definitions when enabled. These help you join wider mesh segments. Apply changes through the Reticulum section in Settings.

## Next steps

- [Navigation and URLs](navigation-and-urls.md) for manual hash entry
- [Using the browser](using-the-browser.md) for favorites and history

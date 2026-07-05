# FAQ

Short answers to common questions.

## What is Ren Browser?

A browser for NomadNet pages on the Reticulum mesh. It is not a general-purpose web browser for the public internet.

## Do I need the internet?

You need Reticulum connectivity to other nodes. That can be entirely offline from the public internet (radio, LAN, etc.).

## Where do I get reticulum-go?

Install and configure [reticulum-go](https://reticulum-go.quad4.io) on the machine that runs Ren Browser. The app uses your Reticulum config (for example under `~/.reticulum-go/`) but does not create your identity or interfaces for you.

## What is NomadNet?

[Nomad Network](https://github.com/markqvist/NomadNet) is an off-grid mesh communications program built on LXMF and Reticulum. Connectable nodes can host pages and files, often written in the Micron markup language. Ren Browser does not embed the NomadNet client. It finds nodes that announce as NomadNet and opens their hosted pages over Reticulum.

## What is Micron?

A bandwidth-efficient markup language used on Nomad Network nodes. Ren Browser renders it to HTML in the content viewer.

## Can I browse normal HTTPS sites?

No. Ren Browser targets Reticulum mesh content, including pages on NomadNet nodes, not arbitrary public web URLs.

## Desktop vs server?

Desktop runs a native window and stores data in local SQLite. Server mode serves the UI over HTTP for use in another browser. See [Server mode](server-mode.md).

## Is server mode safe on the internet?

Not by default. There is no login. Use a VPN, firewall, or authenticated reverse proxy. See [Security](security.md).

## Where is my data?

`~/.renbrowser/renbrowser.db` on desktop by default. See [Data and profiles](data-and-profiles.md).

## How do I verify a release download?

Use `SHA256SUMS.txt` from the release page. See [Security](security.md).

## How do I install an extension?

Settings → Extensions, or unpack into `~/.renbrowser/plugins/<id>/`. See [Extensions](extensions.md).

## How do I type a node address?

Paste the 32-character hex hash or the full `hash:/page/file.mu` URL. See [Navigation and URLs](navigation-and-urls.md).

## Discovery shows nothing

Check Reticulum interfaces in Settings and [Reticulum setup](reticulum-setup.md).

## How do I report a bug?

For security issues use LXMF per [Security](security.md). For code fixes see [Contributing](contributing.md).

## What license is the project?

MIT. Type `license:` in the address bar or read [LICENSE](../../LICENSE).

## What stack is it built with?

Go, Wails v3, Svelte 5, SQLite, and Quad4 Reticulum libraries. See [Architecture](architecture.md).

## Can I run on Android?

Yes when an APK is published for your release. Build from source with the Android SDK if needed. See [Installation](installation.md).

## How do I change keyboard shortcuts?

Settings → Keybinds. See [Keyboard shortcuts](keyboard-shortcuts.md).

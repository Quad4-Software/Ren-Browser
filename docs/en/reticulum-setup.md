# Reticulum setup

Ren Browser uses [Reticulum](https://reticulum.network/) through the `quad4/reticulum-go` stack. This page explains what the app expects and how to fix common mesh issues.

## Default config location

| Item | Default path |
|------|--------------|
| Reticulum config directory | `~/.reticulum-go/` |
| Override flag | `--config /path/to/config` |
| Override env | `REN_BROWSER_CONFIG` or `RETICULUM_CONFIG` |

The exact files inside the directory depend on your Reticulum or reticulum-go setup. Ren Browser starts the stack on launch and reloads interface changes from **Settings**.

## What happens at startup

1. Ren Browser loads your Reticulum config
2. Interfaces come online (UDP, TCP, RNode, and others you configured)
3. Announcements from NomadNet nodes appear in **Discovery**
4. Page requests go out over LXMF and Reticulum to nodes that host Micron pages

If startup fails, check the terminal log (desktop) or container logs (server). The app keeps running so you can still open `about:` and **Settings**.

## Interfaces in Settings

Open **Settings** and find the Reticulum section. You can:

- See which interfaces are active
- View transmit and receive statistics
- Edit the config and apply hot reload without restarting the whole app

Use this when you add a new interface or change keys and want the browser to pick up changes quickly.

## Joining the mesh

You need at least one path to other Reticulum nodes. Common options:

- **Local UDP or TCP** on a LAN with other Reticulum peers
- **RNode** or similar radio hardware
- **Interface definitions** that point at known peers or hubs

Reticulum is outside the scope of this manual. Read the [Reticulum manual](https://reticulum.network/manual/) for interface syntax and identity management.

## NomadNet destinations

NomadNet pages live at Reticulum destinations. In the address bar you can use:

- A full path such as `abcdef0123456789abcdef0123456789:/page/index.mu`
- A bare 32-character hex hash (Ren Browser appends `:/page/index.mu`)

Pages use the Micron markup format. Ren Browser renders them with the built-in Micron pipeline.

## When Discovery is empty

Work through this list:

1. Confirm Reticulum is running inside Ren Browser (Settings shows interfaces)
2. Check that your interfaces match how peers on the mesh are configured
3. Wait a short time after connect. Announcements are not instant
4. Verify you are on the same logical network as nodes you expect to see

## When pages time out or fail

1. Confirm the destination hash is correct
2. Check that you have a route to that destination (not just Discovery visibility)
3. Try another known-good node from Discovery
4. Look at devtools or logs for LXMF or transport errors

## Server and Docker

When you run the `renbrowser` Docker image, mount the host Reticulum directory and run as your host user so the non-root container can read keys and write mesh storage:

```sh
--user "$(id -u):$(id -g)" \
-e HOME=/data \
-v "$HOME/.reticulum-go:/data/.reticulum-go" \
-v "$HOME/.renbrowser:/data/.renbrowser" \
-e REN_BROWSER_CONFIG=/data/.reticulum-go/config
```

Do not mount the config read-only; Reticulum needs to update storage next to the config file.

## Next steps

- [Discovery](discovery.md) for browsing announced nodes
- [Navigation and URLs](navigation-and-urls.md) for address bar formats
- [Troubleshooting](troubleshooting.md) for error messages

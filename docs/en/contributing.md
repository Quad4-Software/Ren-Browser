# Contributing

Ren Browser accepts patches sent over Reticulum. GitHub pull requests may also be welcome depending on project policy. This page follows [CONTRIBUTING.md](../../CONTRIBUTING.md).

## Patch workflow

1. Clone or fork the repository
2. Create a branch and make focused changes
3. Run `task check` when you touch Go or frontend code
4. Commit with a clear message
5. Export patches with `git format-patch`
6. Send the `.patch` file over LXMF

## Sending patches over LXMF

Destination:

```
f489752fbef161c64d65e385a4e9fc74
```

Attach the patch using Sideband, Meshchat, MeshChatX, or any LXMF client that supports file attachments. Include a short description in the message body.

Be patient. Review happens on mesh time.

## Export commands

```sh
# Most recent commit
git format-patch -1

# Last N commits
git format-patch -N

# All commits since main
git format-patch main..HEAD
```

Each commit becomes one `.patch` file.

## Patch guidelines

- One logical change per patch series when possible
- Test before you send
- Match existing code style
- Keep `// SPDX-License-Identifier: MIT` on new Go files
- Disclose AI usage in the message body (see below)

## Licensing

By submitting a patch you agree it is licensed under the [MIT License](../../LICENSE). You confirm you have the right to submit the work.

## Generative AI policy

You may use AI tools if:

- Your setup gives the model enough context
- Your provider does not train on the code you paste

Read [Reticulum Zen](https://reticulum.network/manual/zen.html) and the [Reticulum License](https://reticulum.network/manual/license.html).

**Disclose** which tools you used in the patch message. If you did not use AI in a meaningful way, say so briefly.

Local or offline models are strongly preferred.

You must still read, understand, and test everything you submit. Bulk unreviewed output is not accepted.

## Security issues

Do not send vulnerability details as casual patches on LXMF without coordination. Use the process in [Security](security.md).

## Development setup

See [Development](development.md) for `task dev`, `task check`, and repo layout.

## Next steps

- [Development](development.md)
- [FAQ](faq.md)

# Contributing to Ren Browser

Patches are the preferred way to contribute. Create your changes locally, export a `.patch` file, and send it over Reticulum.

## Generating a patch

1. Clone or fork the repository and make your changes on a branch.

2. Stage and commit your work:

    ```bash
    git add -A
    git commit -m "Short description of the change"
    ```

3. Export the commit(s) as a `.patch` file:

    ```bash
    # Single most recent commit
    git format-patch -1

    # Last N commits
    git format-patch -N

    # All commits since a branch point
    git format-patch main..HEAD
    ```

    This produces one `.patch` file per commit (for example `0001-my-change.patch`).

## Sending the patch

Send the `.patch` file as an LXMF message over Reticulum to:

```
f489752fbef161c64d65e385a4e9fc74
```

You can attach the file using Sideband, Meshchat, MeshChatX, or any LXMF-capable client with attachments support. Include a brief description of what the patch does in the message body.

Lastly, be patient.

## Patch guidelines

- Keep patches focused on a single change or fix.
- Test your changes before exporting. Run `task check` when your change touches Go or frontend code.
- Match existing code style and SPDX headers (`// SPDX-License-Identifier: MIT`).

## Licensing of contributions

By submitting a patch, you agree that your contribution is licensed under the [MIT License](LICENSE), consistent with the per-file SPDX headers in this repository.

You also confirm that you have the right to submit the contribution under these terms (for example, it is your own work, or you have permission from the copyright holder), and that you are not knowingly introducing code under an incompatible license.

## Generative AI policy

You may use generative AI tools when contributing, on the condition that your setup actually supplies the model with enough context to produce sound work and your provider does not train on the code. Read [Reticulum Zen](https://reticulum.network/manual/zen.html) and the [Reticulum License](https://reticulum.network/manual/license.html).

You must disclose AI usage in the patch message body (or commit message, if you prefer): state which tools or services you used in a material way for that change (for example, model or product name, and whether it was local or cloud). If a change was written without meaningful AI assistance, say so briefly. This is so reviewers can judge scope and provenance; it is not a substitute for your own review and testing.

We strongly prefer models that run locally or offline when that is practical for you.

Contributions must still be yours to justify and maintain. Do not submit bulk-generated changes you have not read, understood, and tested. We are not looking for unreviewed AI output or style-only churn from tools used without engineering judgment.

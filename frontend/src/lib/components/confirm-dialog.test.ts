// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it } from "vitest";
import { mount } from "svelte";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";
import ConfirmDialog from "./ConfirmDialog.svelte";

describe("ConfirmDialog component", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
  });

  it("renders nothing when closed", async () => {
    instance = await mountInBody(ConfirmDialog, {
      open: false,
      title: "Confirm",
      message: "Are you sure?",
      onConfirm: () => {},
      onCancel: () => {},
    });
    expect(document.querySelector('[role="alertdialog"]')).toBeNull();
  });

  it("exposes accessible dialog roles and labels", async () => {
    instance = await mountInBody(ConfirmDialog, {
      open: true,
      title: "Reset database",
      message: "This cannot be undone.",
      confirmLabel: "Reset",
      cancelLabel: "Cancel",
      onConfirm: () => {},
      onCancel: () => {},
    });

    const dialog = document.querySelector('[role="alertdialog"]');
    expect(dialog).not.toBeNull();
    expect(dialog?.getAttribute("aria-modal")).toBe("true");
    expect(document.getElementById("confirm-dialog-title")?.textContent).toBe("Reset database");
    expect(document.getElementById("confirm-dialog-message")?.textContent).toContain(
      "cannot be undone",
    );
  });
});

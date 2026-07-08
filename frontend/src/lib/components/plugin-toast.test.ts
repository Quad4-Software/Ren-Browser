// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it } from "vitest";
import { mount } from "svelte";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";
import PluginToast from "./PluginToast.svelte";

describe("PluginToast component", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
  });

  it("renders status toast message", async () => {
    instance = await mountInBody(PluginToast, { message: "Saved", isError: false });
    const toast = document.querySelector(".plugin-toast");
    expect(toast?.getAttribute("role")).toBe("status");
    expect(toast?.textContent).toBe("Saved");
    expect(toast?.classList.contains("error")).toBe(false);
  });

  it("marks error toasts", async () => {
    instance = await mountInBody(PluginToast, { message: "Failed", isError: true });
    expect(document.querySelector(".plugin-toast.error")?.textContent).toBe("Failed");
  });
});

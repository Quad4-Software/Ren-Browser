// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it } from "vitest";
import { mount } from "svelte";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";
import EmptyState from "./EmptyState.svelte";

describe("EmptyState component", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
  });

  it("renders title and description", async () => {
    instance = await mountInBody(EmptyState, {
      title: "Nothing here",
      description: "Try again later",
    });
    expect(document.querySelector(".empty-state .title")?.textContent).toBe("Nothing here");
    expect(document.querySelector(".empty-state .description")?.textContent).toBe(
      "Try again later",
    );
  });
});

// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it, vi } from "vitest";
import { mount } from "svelte";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";
import PageErrorState from "./PageErrorState.svelte";

describe("PageErrorState component", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
  });

  it("renders error copy and retry action", async () => {
    const onRetry = vi.fn();
    instance = await mountInBody(PageErrorState, {
      error: "Node unreachable",
      errorKind: "connection_failed",
      currentURL: "aabbccddeeff00112233445566778899:/page/index.mu",
      onRetry,
    });

    expect(document.querySelector(".error-page")).not.toBeNull();
    expect(document.querySelector(".error-page .url")?.textContent).toContain("aabbccddeeff");
    const retry = document.querySelector(".error-page button.primary");
    expect(retry).not.toBeNull();
    (retry as HTMLButtonElement).click();
    expect(onRetry).toHaveBeenCalled();
  });
});

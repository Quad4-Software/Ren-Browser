// SPDX-License-Identifier: MIT
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./wasm-store", () => ({
  BUNDLED_MICRON_WASM_PARSER_ID: "bundled",
  getMicronWasmParserWasm: vi.fn(async () => null),
  hasStoredMicronWasmParsers: vi.fn(async () => false),
}));

describe("preloadNomadMicronWasm concurrency", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.stubGlobal("WebAssembly", {
      instantiate: vi.fn(async () => {
        throw new Error("not used");
      }),
      instantiateStreaming: vi.fn(async () => {
        throw new Error("not used");
      }),
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response(null, { status: 404 })),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("reuses one in-flight promise for the same parser id", async () => {
    const { preloadNomadMicronWasm, invalidateNomadMicronWasmPreload } =
      await import("./wasm-loader");
    invalidateNomadMicronWasmPreload();

    const first = preloadNomadMicronWasm("bundled");
    const second = preloadNomadMicronWasm("bundled");
    expect(second).toBe(first);

    await expect(first).resolves.toBe(false);
  });

  it("does not invalidate an in-flight same-id load on a second call", async () => {
    const { preloadNomadMicronWasm, invalidateNomadMicronWasmPreload } =
      await import("./wasm-loader");
    invalidateNomadMicronWasmPreload();

    let fetchCount = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        fetchCount += 1;
        await new Promise((resolve) => setTimeout(resolve, 20));
        return new Response(null, { status: 404 });
      }),
    );

    const first = preloadNomadMicronWasm("bundled");
    const second = preloadNomadMicronWasm("bundled");
    expect(second).toBe(first);
    await Promise.allSettled([first, second]);
    // One shared load attempt, not two competing invalidating loads.
    expect(fetchCount).toBeGreaterThan(0);
    expect(fetchCount).toBeLessThan(4);
  });
});

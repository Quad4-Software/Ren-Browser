// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  LARGE_MICRON_RAW_BYTES,
  micronRendererBadgeLabel,
  normalizeMicronRendererPreference,
  nodeHashFromURL,
  pagePathFromURL,
  resolveEffectiveMicronEngine,
  shouldPreloadMicronWasm,
  usesClientMicronRenderer,
} from "./render-page";

describe("micron render-page helpers", () => {
  it("detects client renderers", () => {
    expect(usesClientMicronRenderer("go")).toBe(false);
    expect(usesClientMicronRenderer("js")).toBe(true);
    expect(usesClientMicronRenderer("wasm")).toBe(true);
  });

  it("parses node hash and path from mesh urls", () => {
    const url = "abb3ebcd03cb2388a838e70c001291f9:/page/index.mu";
    expect(nodeHashFromURL(url)).toBe("abb3ebcd03cb2388a838e70c001291f9");
    expect(pagePathFromURL(url)).toBe("/page/index.mu");
  });

  it("defaults preference to auto", () => {
    expect(normalizeMicronRendererPreference(undefined)).toBe("auto");
    expect(normalizeMicronRendererPreference("wasm")).toBe("wasm");
  });

  it("resolves auto to go when server html is present", () => {
    const ctx = {
      wasmEnabled: true,
      wasmAvailable: true,
      wasmReady: true,
      hasServerHtml: true,
    };
    expect(resolveEffectiveMicronEngine("auto", ctx)).toBe("go");

    expect(
      resolveEffectiveMicronEngine("auto", {
        ...ctx,
        hasServerHtml: false,
      }),
    ).toBe("wasm");

    expect(
      resolveEffectiveMicronEngine("auto", {
        ...ctx,
        hasServerHtml: false,
        wasmReady: false,
      }),
    ).toBe("js");
  });

  it("forces go for large pages when server html exists", () => {
    expect(
      resolveEffectiveMicronEngine("wasm", {
        wasmEnabled: true,
        wasmAvailable: true,
        wasmReady: true,
        hasServerHtml: true,
        rawBytes: LARGE_MICRON_RAW_BYTES,
      }),
    ).toBe("go");

    expect(
      resolveEffectiveMicronEngine("auto", {
        wasmEnabled: true,
        wasmAvailable: true,
        wasmReady: true,
        hasServerHtml: true,
        rawBytes: LARGE_MICRON_RAW_BYTES + 1,
      }),
    ).toBe("go");

    expect(
      resolveEffectiveMicronEngine("js", {
        wasmEnabled: true,
        wasmAvailable: true,
        wasmReady: true,
        hasServerHtml: true,
        rawBytes: LARGE_MICRON_RAW_BYTES,
      }),
    ).toBe("js");
  });

  it("does not force go for large pages without server html", () => {
    expect(
      resolveEffectiveMicronEngine("wasm", {
        wasmEnabled: true,
        wasmAvailable: true,
        wasmReady: true,
        hasServerHtml: false,
        rawBytes: LARGE_MICRON_RAW_BYTES,
      }),
    ).toBe("wasm");
  });

  it("keeps explicit manual preferences for small pages", () => {
    expect(
      resolveEffectiveMicronEngine("go", {
        wasmEnabled: true,
        wasmAvailable: true,
        wasmReady: true,
        hasServerHtml: true,
      }),
    ).toBe("go");

    expect(
      resolveEffectiveMicronEngine("js", {
        wasmEnabled: true,
        wasmAvailable: true,
        wasmReady: true,
        hasServerHtml: true,
      }),
    ).toBe("js");

    expect(
      resolveEffectiveMicronEngine("wasm", {
        wasmEnabled: true,
        wasmAvailable: true,
        wasmReady: true,
        hasServerHtml: true,
        rawBytes: 1024,
      }),
    ).toBe("wasm");
  });

  it("preloads wasm for auto and wasm preferences", () => {
    expect(shouldPreloadMicronWasm("auto", true)).toBe(true);
    expect(shouldPreloadMicronWasm("wasm", true)).toBe(true);
    expect(shouldPreloadMicronWasm("go", true)).toBe(false);
    expect(shouldPreloadMicronWasm("auto", false)).toBe(false);
  });

  it("falls back from wasm preference when wasm is not ready yet", () => {
    expect(
      resolveEffectiveMicronEngine("wasm", {
        wasmEnabled: true,
        wasmAvailable: true,
        wasmReady: false,
        hasServerHtml: false,
      }),
    ).toBe("js");

    expect(
      resolveEffectiveMicronEngine("wasm", {
        wasmEnabled: true,
        wasmAvailable: true,
        wasmReady: false,
        hasServerHtml: true,
      }),
    ).toBe("go");
  });
});

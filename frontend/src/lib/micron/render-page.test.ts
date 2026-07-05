import { describe, expect, it } from "vitest";
import {
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

  it("resolves auto fallback chain wasm -> go -> js", () => {
    const ctx = {
      wasmEnabled: true,
      wasmAvailable: true,
      wasmReady: true,
      hasServerHtml: true,
    };
    expect(resolveEffectiveMicronEngine("auto", ctx)).toBe("wasm");

    expect(resolveEffectiveMicronEngine("auto", { ...ctx, wasmReady: false })).toBe("go");

    expect(
      resolveEffectiveMicronEngine("auto", {
        ...ctx,
        wasmReady: false,
        hasServerHtml: false,
      }),
    ).toBe("js");
  });

  it("keeps explicit manual preferences", () => {
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
  });

  it("preloads wasm for auto and wasm preferences", () => {
    expect(shouldPreloadMicronWasm("auto", true)).toBe(true);
    expect(shouldPreloadMicronWasm("wasm", true)).toBe(true);
    expect(shouldPreloadMicronWasm("go", true)).toBe(false);
    expect(shouldPreloadMicronWasm("auto", false)).toBe(false);
  });

  it("labels auto badge with effective renderer", () => {
    expect(micronRendererBadgeLabel("auto", "go")).toBe("Auto: Micron Go");
    expect(micronRendererBadgeLabel("wasm", "wasm", "Bundled")).toBe("Micron WASM (Bundled)");
  });
});

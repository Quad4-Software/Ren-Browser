import { describe, expect, it, vi } from "vitest";
import {
  parsePluginI18nRef,
  resolvePluginLabel,
  loadPluginCatalogs,
  clearPluginI18n,
} from "./plugin-i18n";

describe("plugin i18n", () => {
  it("parses manifest title references", () => {
    expect(parsePluginI18nRef("%panels.translator%")).toBe("panels.translator");
    expect(parsePluginI18nRef("Micron Translator")).toBeNull();
  });

  it("resolves labels from loaded catalogs", async () => {
    clearPluginI18n();
    const pluginId = "demo.plugin";
    const fetchMock = vi.fn(async (url: string) => {
      if (url.endsWith("/locales/en.json")) {
        return {
          ok: true,
          json: async () => ({ panels: { translator: "Translator panel" } }),
        };
      }
      return { ok: false };
    });
    vi.stubGlobal("fetch", fetchMock);

    await loadPluginCatalogs(pluginId, "en");
    expect(resolvePluginLabel(pluginId, "%panels.translator%", "en")).toBe("Translator panel");
    expect(resolvePluginLabel(pluginId, "Literal title", "en")).toBe("Literal title");
    expect(resolvePluginLabel(pluginId, "%panels.missing%", "de")).toBe("Missing");

    vi.unstubAllGlobals();
    clearPluginI18n();
  });
});

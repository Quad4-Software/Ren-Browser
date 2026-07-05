// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { panelKey, parsePanelKey, setContributions } from "./registry.js";

describe("plugin registry", () => {
  it("builds and parses panel keys", () => {
    const key = panelKey("renbrowser.hello", "hello");
    expect(key).toBe("plugin:renbrowser.hello:hello");
    expect(parsePanelKey(key)).toEqual({ pluginId: "renbrowser.hello", panelId: "hello" });
  });

  it("stores contributions snapshot", () => {
    setContributions({
      panels: [{ pluginId: "a", id: "p", title: "P", entry: "main.js" }],
      commands: [],
      devtools: [],
      urlSchemes: [],
    });
    expect(parsePanelKey("plugin:a:p")).not.toBeNull();
  });
});

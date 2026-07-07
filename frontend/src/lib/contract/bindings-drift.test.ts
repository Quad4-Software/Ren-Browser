// SPDX-License-Identifier: MIT
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import * as browserService from "../../../bindings/renbrowser/internal/app/browserservice.js";
import * as pluginHost from "../../../bindings/renbrowser/internal/app/pluginhost.js";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../../..");

function bindingExports(relPath: string): Set<string> {
  const source = readFileSync(path.join(repoRoot, relPath), "utf8");
  const names = new Set<string>();
  for (const match of source.matchAll(/^export function (\w+)\(/gm)) {
    names.add(match[1]);
  }
  return names;
}

describe("bindings contract drift", () => {
  it("exports BrowserService methods used by the shell", () => {
    const required = [
      "Navigate",
      "GetBrowserPrefs",
      "SetBrowserPrefs",
      "FetchCommunityInterfaces",
      "ImportCommunityInterfaces",
      "ListIdentities",
      "CreateIdentity",
      "SetActiveIdentity",
      "GetTabs",
      "SaveTabs",
    ] as const;

    for (const name of required) {
      expect(typeof (browserService as Record<string, unknown>)[name]).toBe("function");
    }
  });

  it("matches generated BrowserService binding exports", () => {
    const generated = bindingExports("frontend/bindings/renbrowser/internal/app/browserservice.ts");
    for (const name of Object.keys(browserService)) {
      if (name.startsWith("$")) {
        continue;
      }
      expect(generated.has(name)).toBe(true);
    }
  });

  it("matches generated PluginHost binding exports", () => {
    const generated = bindingExports("frontend/bindings/renbrowser/internal/app/pluginhost.ts");
    for (const name of Object.keys(pluginHost)) {
      if (name.startsWith("$")) {
        continue;
      }
      expect(generated.has(name)).toBe(true);
    }
  });
});

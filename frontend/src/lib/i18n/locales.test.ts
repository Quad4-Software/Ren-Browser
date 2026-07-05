// SPDX-License-Identifier: MIT
import { readdirSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { flattenMessageKeys, listLocaleCatalogs } from "./catalog";
import { DEFAULT_LOCALE, SUPPORTED_LOCALES } from "./locales";

const localesDir = join(dirname(fileURLToPath(import.meta.url)), "locales");

function loadLocaleFiles(): Record<string, Record<string, unknown>> {
  const files = readdirSync(localesDir).filter(
    (name) => name.endsWith(".json") && !name.startsWith("_"),
  );
  const locales: Record<string, Record<string, unknown>> = {};
  for (const file of files) {
    const code = file.replace(/\.json$/, "");
    locales[code] = JSON.parse(readFileSync(join(localesDir, file), "utf8")) as Record<
      string,
      unknown
    >;
  }
  return locales;
}

describe("locale catalogs", () => {
  const baseKeys = flattenMessageKeys(listLocaleCatalogs()[DEFAULT_LOCALE]);
  const localeFiles = loadLocaleFiles();

  it("includes every supported locale file", () => {
    for (const entry of SUPPORTED_LOCALES) {
      expect(localeFiles[entry.code], `missing locales/${entry.code}.json`).toBeDefined();
    }
  });

  for (const entry of SUPPORTED_LOCALES) {
    if (entry.code === DEFAULT_LOCALE) {
      continue;
    }

    it(`${entry.code}.json has the same keys as ${DEFAULT_LOCALE}.json`, () => {
      const keys = flattenMessageKeys(localeFiles[entry.code] as never);
      expect(keys).toEqual(baseKeys);
    });

    it(`${entry.code}.json has no empty translation values`, () => {
      const keys = flattenMessageKeys(localeFiles[entry.code] as never);
      for (const key of keys) {
        const parts = key.split(".");
        let current: unknown = localeFiles[entry.code];
        for (const part of parts) {
          current = (current as Record<string, unknown>)[part];
        }
        expect(current, `empty value for ${entry.code}.${key}`).not.toBe("");
      }
    });
  }

  it("_template.json mirrors base keys with empty placeholders", () => {
    const template = JSON.parse(
      readFileSync(join(localesDir, "_template.json"), "utf8"),
    ) as Record<string, unknown>;
    const templateKeys = flattenMessageKeys(template as never);
    expect(templateKeys).toEqual(baseKeys);
    for (const key of templateKeys) {
      const parts = key.split(".");
      let current: unknown = template;
      for (const part of parts) {
        current = (current as Record<string, unknown>)[part];
      }
      expect(current, `template value for ${key}`).toBe("");
    }
  });
});

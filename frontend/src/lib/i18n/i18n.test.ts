// SPDX-License-Identifier: MIT
import { beforeEach, describe, expect, it } from "vitest";
import { getCatalogLocale, setCatalogLocale, translate, translatePermission } from "./catalog";
import { DEFAULT_LOCALE } from "./locales";

describe("translate", () => {
  beforeEach(() => {
    setCatalogLocale(DEFAULT_LOCALE);
  });

  it("resolves nested keys", () => {
    expect(translate("chrome.back")).toBe("Back");
  });

  it("interpolates parameters", () => {
    expect(translate("common.entries", { current: 3, max: 128 })).toBe("3 / 128 entries");
  });

  it("falls back to English for missing keys in other locales", () => {
    setCatalogLocale("de");
    expect(translate("chrome.back")).toBe("Zurück");
    setCatalogLocale("xx");
    expect(getCatalogLocale()).toBe(DEFAULT_LOCALE);
  });

  it("translates dotted extension permission ids", () => {
    expect(translatePermission("storage.plugin")).toBe("Store data for this extension");
    expect(translatePermission("network.fetch")).toBe("Make outbound network requests");
    expect(translatePermission("unknown.permission")).toBe("unknown.permission");
  });
});

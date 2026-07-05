// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  isAllowedNavigationURL,
  isBlockedExternalHref,
  isBlockedNavigationURL,
} from "./navigation-guard";

describe("isBlockedExternalHref", () => {
  it("blocks common internet schemes", () => {
    expect(isBlockedExternalHref("https://example.com")).toBe(true);
    expect(isBlockedExternalHref("http://example.com/path")).toBe(true);
    expect(isBlockedExternalHref("//cdn.example.com/x")).toBe(true);
    expect(isBlockedExternalHref("javascript:alert(1)")).toBe(true);
    expect(isBlockedExternalHref("mailto:test@example.com")).toBe(true);
  });

  it("allows mesh and fragment links", () => {
    expect(isBlockedExternalHref("#section")).toBe(false);
    expect(isBlockedExternalHref("/page/index.mu")).toBe(false);
    expect(
      isBlockedExternalHref("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu"),
    ).toBe(false);
    expect(isBlockedExternalHref("hello:world")).toBe(false);
  });
});

describe("isBlockedNavigationURL", () => {
  it("blocks web urls", () => {
    expect(isBlockedNavigationURL("https://example.com")).toBe(true);
    expect(isBlockedNavigationURL("ftp://files.example.com")).toBe(true);
  });

  it("allows app and mesh urls", () => {
    expect(isBlockedNavigationURL("about:")).toBe(false);
    expect(isBlockedNavigationURL("docs:?lang=en")).toBe(false);
    expect(
      isBlockedNavigationURL("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu"),
    ).toBe(false);
    expect(isBlockedNavigationURL("rns://abb3ebcd03cb2388a838e70c001291f9/page/index.mu")).toBe(
      false,
    );
    expect(isBlockedNavigationURL("hello:panel")).toBe(false);
  });
});

describe("isAllowedNavigationURL", () => {
  it("rejects blocked urls", () => {
    expect(isAllowedNavigationURL("https://example.com")).toBe(false);
    expect(isAllowedNavigationURL("")).toBe(false);
  });
});

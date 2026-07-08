// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { unwrapDeepLink } from "./deeplink";

const hash = "abb3ebcd03cb2388a838e70c001291f9";

describe("unwrapDeepLink", () => {
  it("unwraps renbrowser wrappers", () => {
    expect(unwrapDeepLink("renbrowser:about")).toBe("about:");
    expect(unwrapDeepLink("renbrowser://about")).toBe("about:");
    expect(unwrapDeepLink("renbrowser://open?url=about%3A")).toBe("about:");
    expect(unwrapDeepLink(`renbrowser://${hash}/page/index.mu`)).toBe(`${hash}:/page/index.mu`);
    expect(unwrapDeepLink(`renbrowser://rns/${hash}/page/home.mu`)).toBe(
      `rns://${hash}/page/home.mu`,
    );
  });

  it("passes through internal urls", () => {
    expect(unwrapDeepLink("about:")).toBe("about:");
    expect(unwrapDeepLink(`rns://${hash}/page/x.mu`)).toBe(`rns://${hash}/page/x.mu`);
    expect(unwrapDeepLink(`${hash}:/page/index.mu`)).toBe(`${hash}:/page/index.mu`);
  });

  it("rejects external and dangerous schemes", () => {
    expect(unwrapDeepLink("https://example.com")).toBe("");
    expect(unwrapDeepLink("javascript:alert(1)")).toBe("");
    expect(unwrapDeepLink("renbrowser://open?url=https%3A%2F%2Fevil.test")).toBe("");
    expect(unwrapDeepLink("data:text/html,hi")).toBe("");
    expect(unwrapDeepLink("file:///etc/passwd")).toBe("");
  });

  it("handles edge cases", () => {
    expect(unwrapDeepLink("")).toBe("");
    expect(unwrapDeepLink("   ")).toBe("");
    expect(unwrapDeepLink("renbrowser:")).toBe("");
    expect(unwrapDeepLink("renbrowser://open?url=")).toBe("");
    expect(unwrapDeepLink("about:\0x")).toBe("");
    expect(unwrapDeepLink("  RenBrowser://License  ")).toBe("license:");
  });
});

// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { findFavoriteUrl, isFavoriteUrl, isFavoritableUrl } from "./favorites";

const HASH_A = "abb3ebcd03cb2388a838e70c001291f9";
const HASH_B = "deadbeefdeadbeefdeadbeefdeadbeef";

describe("findFavoriteUrl", () => {
  it("returns undefined for empty input or empty favorites", () => {
    expect(findFavoriteUrl("", ["a:/b"])).toBeUndefined();
    expect(findFavoriteUrl("   ", ["a:/b"])).toBeUndefined();
    expect(findFavoriteUrl(`${HASH_A}:/page/index.mu`, [])).toBeUndefined();
  });

  it("prefers an exact stored match", () => {
    const favorites = [`${HASH_A}:/page/index.mu`, `${HASH_A}:/page/about.mu`];
    expect(findFavoriteUrl(`${HASH_A}:/page/about.mu`, favorites)).toBe(`${HASH_A}:/page/about.mu`);
  });

  it("matches hash-equivalent entries like the Discovery panel", () => {
    const favorites = [`${HASH_A}:/page/index.mu`];
    expect(findFavoriteUrl(HASH_A, favorites)).toBe(`${HASH_A}:/page/index.mu`);
    expect(findFavoriteUrl(`${HASH_A}:/page/index.mu`, favorites)).toBe(`${HASH_A}:/page/index.mu`);
    expect(findFavoriteUrl(`${HASH_A}:/page/deep/page.mu`, favorites)).toBe(
      `${HASH_A}:/page/index.mu`,
    );
  });

  it("matches a bare-hash favorite against full mesh urls", () => {
    const favorites = [HASH_A];
    expect(findFavoriteUrl(`${HASH_A}:/page/index.mu`, favorites)).toBe(HASH_A);
  });

  it("compares node hashes case-insensitively", () => {
    const favorites = [`${HASH_A.toUpperCase()}:/page/index.mu`];
    expect(findFavoriteUrl(`${HASH_A}:/page/index.mu`, favorites)).toBe(
      `${HASH_A.toUpperCase()}:/page/index.mu`,
    );
  });

  it("does not match a different node", () => {
    const favorites = [`${HASH_A}:/page/index.mu`];
    expect(findFavoriteUrl(`${HASH_B}:/page/index.mu`, favorites)).toBeUndefined();
  });

  it("does not treat non-mesh schemes as hash-equivalent", () => {
    const favorites = ["document:/book.epub", "document:/other.epub"];
    expect(findFavoriteUrl("document:/book.epub", favorites)).toBe("document:/book.epub");
    expect(findFavoriteUrl("document:/third.epub", favorites)).toBeUndefined();
    expect(findFavoriteUrl("about:", ["about:"])).toBe("about:");
    expect(findFavoriteUrl("docs:?page=faq", ["docs:"])).toBeUndefined();
  });
});

describe("isFavoriteUrl", () => {
  it("reflects findFavoriteUrl", () => {
    const favorites = [`${HASH_A}:/page/index.mu`];
    expect(isFavoriteUrl(`${HASH_A}:/page/index.mu`, favorites)).toBe(true);
    expect(isFavoriteUrl(`${HASH_B}:/page/index.mu`, favorites)).toBe(false);
    expect(isFavoriteUrl("", favorites)).toBe(false);
  });
});

describe("isFavoritableUrl", () => {
  it("accepts mesh pages and document urls", () => {
    expect(isFavoritableUrl(`${HASH_A}:/page/index.mu`)).toBe(true);
    expect(isFavoritableUrl("document:/book.epub")).toBe(true);
  });

  it("rejects built-in app pages", () => {
    for (const url of [
      "about:",
      "license:",
      "editor:",
      "config:",
      "settings:",
      "docs:",
      "docs:?lang=en&page=faq",
      "docs?lang=en",
    ]) {
      expect(isFavoritableUrl(url), url).toBe(false);
    }
  });

  it("rejects empty input", () => {
    expect(isFavoritableUrl("")).toBe(false);
    expect(isFavoritableUrl("   ")).toBe(false);
  });
});

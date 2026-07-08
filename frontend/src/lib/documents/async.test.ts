// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { DocumentTimeoutError, documentErrorI18nKey, withTimeout } from "./async";

describe("document async", () => {
  it("resolves before timeout", async () => {
    await expect(withTimeout(Promise.resolve(42), 50, "test")).resolves.toBe(42);
  });

  it("rejects on timeout", async () => {
    await expect(withTimeout(new Promise(() => {}), 20, "EPUB parse")).rejects.toBeInstanceOf(
      DocumentTimeoutError,
    );
  });

  it("maps corrupted zip errors to i18n key", () => {
    expect(
      documentErrorI18nKey(new Error("Corrupted zip: can't find end of central directory")),
    ).toBe("documents.corruptEpub");
  });
});

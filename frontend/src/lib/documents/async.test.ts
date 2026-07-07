// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { DocumentTimeoutError, withTimeout } from "./async";

describe("document async", () => {
  it("resolves before timeout", async () => {
    await expect(withTimeout(Promise.resolve(42), 50, "test")).resolves.toBe(42);
  });

  it("rejects on timeout", async () => {
    await expect(
      withTimeout(new Promise(() => {}), 20, "EPUB parse"),
    ).rejects.toBeInstanceOf(DocumentTimeoutError);
  });
});

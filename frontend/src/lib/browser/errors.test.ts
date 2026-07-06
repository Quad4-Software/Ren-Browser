// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { normalizePageErrorKind, pageErrorContent } from "./errors";

describe("errors", () => {
  it("maps reticulum readiness failures to connection_failed", () => {
    expect(normalizePageErrorKind(undefined, "reticulum not ready")).toBe("connection_failed");
  });

  it("maps empty errors to unknown", () => {
    expect(normalizePageErrorKind(undefined, "")).toBe("unknown");
  });

  it("uses detail text for internal errors", () => {
    const page = pageErrorContent("internal", "disk full");
    expect(page.description).toBe("disk full");
    expect(page.showRetry).toBe(true);
  });

  it("maps oversized responses to payload_too_large", () => {
    expect(
      normalizePageErrorKind(undefined, "response too large: received 16 bytes (limit 8)"),
    ).toBe("payload_too_large");
  });

  it("enables database reset for corrupt store errors", () => {
    const page = pageErrorContent("database_corrupt", "");
    expect(page.showResetDatabase).toBe(true);
    expect(page.showRetry).toBe(false);
  });
});

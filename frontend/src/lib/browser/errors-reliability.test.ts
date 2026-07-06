// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { normalizePageErrorKind, pageErrorContent } from "./errors";

describe("errors reliability regressions", () => {
  it("honors backend payload_too_large kind", () => {
    expect(normalizePageErrorKind("payload_too_large", "")).toBe("payload_too_large");
  });

  it("renders payload_too_large without retry", () => {
    const page = pageErrorContent(
      "payload_too_large",
      "response too large: received 16 bytes (limit 8)",
    );
    expect(page.showRetry).toBe(false);
    expect(page.description).toContain("16 bytes");
    expect(page.tone).toBe("danger");
  });

  it("maps shutdown style failures to internal", () => {
    expect(normalizePageErrorKind(undefined, "application shutting down")).toBe("internal");
  });
});

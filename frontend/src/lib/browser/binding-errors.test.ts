// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  formatBindingError,
  parseWailsRuntimeError,
  unwrapBindingErrorMessage,
} from "./binding-errors";

describe("parseWailsRuntimeError", () => {
  it("parses wails runtime error payloads", () => {
    const raw = JSON.stringify({
      message: "requires renbrowser >=0.2.0 (running 0.1.0)",
      cause: {},
      kind: "RuntimeError",
    });
    expect(parseWailsRuntimeError(raw)?.message).toBe(
      "requires renbrowser >=0.2.0 (running 0.1.0)",
    );
  });

  it("decodes escaped characters in messages", () => {
    const raw =
      '{"message":"requires renbrowser \\u003e=0.2.0 (running 0.1.0)","cause":{},"kind":"RuntimeError"}';
    expect(unwrapBindingErrorMessage(raw)).toBe("requires renbrowser >=0.2.0 (running 0.1.0)");
  });
});

describe("formatBindingError", () => {
  it("unwraps runtime errors thrown as Error objects", () => {
    const raw =
      '{"message":"requires renbrowser >=0.2.0 (running 0.1.0)","cause":{},"kind":"RuntimeError"}';
    expect(formatBindingError(new Error(raw))).toBe("requires renbrowser >=0.2.0 (running 0.1.0)");
  });

  it("returns plain backend messages unchanged", () => {
    expect(formatBindingError(new Error("permission denied"))).toBe("permission denied");
  });

  it("uses the fallback for empty errors", () => {
    expect(formatBindingError("", "fallback")).toBe("fallback");
    expect(formatBindingError(undefined, "fallback")).toBe("fallback");
  });
});

// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  formatBindingError,
  humanizeBindingError,
  isWailsRuntimeErrorPayload,
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

describe("isWailsRuntimeErrorPayload", () => {
  it("detects structured runtime errors", () => {
    expect(
      isWailsRuntimeErrorPayload({
        message: "plugin renbrowser.micron-translator permission network.fetch not granted",
        cause: {},
        kind: "RuntimeError",
      }),
    ).toBe(true);
  });
});

describe("humanizeBindingError", () => {
  it("rewrites permission errors into readable text", () => {
    const message = humanizeBindingError(
      "plugin renbrowser.micron-translator permission network.fetch not granted",
    );
    expect(message).toContain("Micron Translator");
    expect(message).toContain("Make outbound network requests");
    expect(message).not.toContain("{");
    expect(message).not.toContain("RuntimeError");
  });
});

describe("formatBindingError", () => {
  it("unwraps runtime errors thrown as Error objects", () => {
    const raw =
      '{"message":"requires renbrowser >=0.2.0 (running 0.1.0)","cause":{},"kind":"RuntimeError"}';
    expect(formatBindingError(new Error(raw))).toBe("requires renbrowser >=0.2.0 (running 0.1.0)");
  });

  it("unwraps structured runtime error objects", () => {
    expect(
      formatBindingError({
        message: "plugin renbrowser.micron-translator permission network.fetch not granted",
        cause: {},
        kind: "RuntimeError",
      }),
    ).toContain("Micron Translator");
  });

  it("returns plain backend messages unchanged", () => {
    expect(formatBindingError(new Error("permission denied"))).toBe("permission denied");
  });

  it("humanizes identity sentinel errors", () => {
    expect(humanizeBindingError("identity not found")).toContain("not found");
    expect(humanizeBindingError("identity already exists: abc")).toContain("already");
    expect(humanizeBindingError("identity file selection canceled")).toContain("canceled");
  });

  it("uses the fallback for empty errors", () => {
    expect(formatBindingError("", "fallback")).toBe("fallback");
    expect(formatBindingError(undefined, "fallback")).toBe("fallback");
  });
});

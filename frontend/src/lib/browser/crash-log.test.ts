import { describe, expect, it } from "vitest";
import { buildCrashDebugLog, crashErrorMessage } from "./crash-log";

describe("crash log", () => {
  it("formats error messages", () => {
    expect(crashErrorMessage(new Error("boom"))).toBe("boom");
    expect(crashErrorMessage("plain")).toBe("plain");
  });

  it("builds a debug report with stack traces", () => {
    const err = new Error("render failed");
    const report = buildCrashDebugLog("render failed", err);
    expect(report).toContain("Ren Browser crash report");
    expect(report).toContain("Error: render failed");
    expect(report).toContain("Stack:");
  });
});

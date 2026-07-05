import { describe, expect, it } from "vitest";
import {
  BUNDLED_MICRON_WASM_PARSER_ID,
  normalizeMicronWasmParserId,
  parseShasums256ForFilename,
} from "./wasm-store";

describe("wasm-store helpers", () => {
  it("parses shasums lines", () => {
    const hex = "a".repeat(64);
    const text = `${hex}  micron-parser-go.wasm\n`;
    expect(parseShasums256ForFilename(text, "micron-parser-go.wasm")).toBe(hex);
  });

  it("normalizes parser id against available parsers", () => {
    const available = new Set([BUNDLED_MICRON_WASM_PARSER_ID, "custom-1"]);
    expect(normalizeMicronWasmParserId("custom-1", available)).toBe("custom-1");
    expect(normalizeMicronWasmParserId("missing", available)).toBe(BUNDLED_MICRON_WASM_PARSER_ID);
  });
});

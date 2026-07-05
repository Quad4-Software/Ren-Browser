// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it, vi } from "vitest";
import { randomId } from "./id";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("randomId", () => {
  it("uses crypto.randomUUID when available", () => {
    const randomUUID = vi.fn(() => "550e8400-e29b-41d4-a716-446655440000");
    vi.stubGlobal("crypto", { randomUUID });
    expect(randomId()).toBe("550e8400-e29b-41d4-a716-446655440000");
    expect(randomUUID).toHaveBeenCalledOnce();
  });

  it("falls back when randomUUID throws", () => {
    const randomUUID = vi.fn(() => {
      throw new DOMException("Not allowed", "SecurityError");
    });
    vi.stubGlobal("crypto", { randomUUID });
    const id = randomId();
    expect(id).toMatch(/^ren-/);
    expect(randomUUID).toHaveBeenCalledOnce();
  });

  it("falls back when crypto.randomUUID is missing", () => {
    vi.stubGlobal("crypto", {});
    const a = randomId();
    const b = randomId();
    expect(a).toMatch(/^ren-/);
    expect(b).toMatch(/^ren-/);
    expect(a).not.toBe(b);
  });

  it("falls back when crypto is undefined", () => {
    vi.stubGlobal("crypto", undefined);
    expect(randomId()).toMatch(/^ren-/);
  });
});

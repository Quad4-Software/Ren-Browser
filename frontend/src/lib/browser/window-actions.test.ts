// SPDX-License-Identifier: MIT
import { describe, expect, it, vi } from "vitest";
import { handleTitlebarDoubleClick, type WindowRuntime } from "./window-actions";

function mockWindow(state: {
  maximised?: boolean;
  minimised?: boolean;
  fail?: keyof WindowRuntime;
}): WindowRuntime {
  let maximised = !!state.maximised;
  let minimised = !!state.minimised;

  const fail = (method: keyof WindowRuntime) => {
    if (state.fail === method) {
      throw new Error(`${method} failed`);
    }
  };

  return {
    IsMaximised: vi.fn(async () => {
      fail("IsMaximised");
      return maximised;
    }),
    IsMinimised: vi.fn(async () => {
      fail("IsMinimised");
      return minimised;
    }),
    Maximise: vi.fn(async () => {
      fail("Maximise");
      maximised = true;
      minimised = false;
    }),
    UnMaximise: vi.fn(async () => {
      fail("UnMaximise");
      maximised = false;
    }),
    UnMinimise: vi.fn(async () => {
      fail("UnMinimise");
      minimised = false;
    }),
  };
}

describe("handleTitlebarDoubleClick", () => {
  it("maximizes a normal window", async () => {
    const api = mockWindow({});
    const result = await handleTitlebarDoubleClick(api);
    expect(result).toEqual({ ok: true, action: "maximized" });
    expect(api.Maximise).toHaveBeenCalledTimes(1);
    expect(api.UnMaximise).not.toHaveBeenCalled();
  });

  it("restores a maximized window", async () => {
    const api = mockWindow({ maximised: true });
    const result = await handleTitlebarDoubleClick(api);
    expect(result).toEqual({ ok: true, action: "restored" });
    expect(api.UnMaximise).toHaveBeenCalledTimes(1);
    expect(api.Maximise).not.toHaveBeenCalled();
  });

  it("unminimizes then maximizes a minimized window", async () => {
    const api = mockWindow({ minimised: true });
    const result = await handleTitlebarDoubleClick(api);
    expect(result).toEqual({ ok: true, action: "maximized" });
    expect(api.UnMinimise).toHaveBeenCalledTimes(1);
    expect(api.Maximise).toHaveBeenCalledTimes(1);
  });

  it("returns an error when the window API fails", async () => {
    const api = mockWindow({ fail: "Maximise" });
    const result = await handleTitlebarDoubleClick(api);
    expect(result.ok).toBe(false);
    expect(result.error).toContain("Maximise failed");
  });
});

// SPDX-License-Identifier: MIT
import { describe, expect, it, vi } from "vitest";
import {
  clampSidebarWidth,
  clampSplitRatio,
  createDragScheduler,
  SIDEBAR_MIN_PX,
  sidebarWidthFromPointer,
  splitRatioFromPointer,
} from "./pane-resize";

function rect(width: number, height = 400): DOMRect {
  return {
    x: 0,
    y: 0,
    left: 0,
    top: 0,
    right: width,
    bottom: height,
    width,
    height,
    toJSON() {
      return {};
    },
  };
}

function pointer(clientX: number, clientY = 0): PointerEvent {
  return { clientX, clientY } as PointerEvent;
}

describe("pane resize math", () => {
  it("clamps split ratios to the allowed range", () => {
    expect(clampSplitRatio(10)).toBe(25);
    expect(clampSplitRatio(90)).toBe(75);
    expect(clampSplitRatio(52)).toBe(52);
    expect(clampSplitRatio(Number.NaN)).toBe(25);
  });

  it("clamps sidebar width against the container", () => {
    expect(clampSidebarWidth(100, 1200)).toBe(SIDEBAR_MIN_PX);
    expect(clampSidebarWidth(900, 1200)).toBe(560);
    expect(clampSidebarWidth(400, 700)).toBe(350);
  });

  it("maps a pointer to a split ratio", () => {
    expect(splitRatioFromPointer(pointer(400), rect(1000))).toBe(40);
    expect(splitRatioFromPointer(pointer(50), rect(1000))).toBe(25);
  });

  it("maps a pointer to a right-side sidebar width", () => {
    expect(sidebarWidthFromPointer(pointer(800), rect(1200))).toBe(400);
    expect(sidebarWidthFromPointer(pointer(1190), rect(1200))).toBe(SIDEBAR_MIN_PX);
  });

  it("coalesces drag frames to the latest pointer event", () => {
    const applied: number[] = [];
    vi.stubGlobal("requestAnimationFrame", (cb: FrameRequestCallback) => {
      cb(0);
      return 1;
    });
    vi.stubGlobal("cancelAnimationFrame", () => {});
    const drag = createDragScheduler((event) => {
      applied.push(event.clientX);
    });
    drag.move(pointer(10));
    expect(applied).toEqual([10]);
    vi.unstubAllGlobals();
  });
});

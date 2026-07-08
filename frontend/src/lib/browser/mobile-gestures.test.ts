// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  clampBackOffset,
  clampPullOffset,
  isBackEdgeStart,
  isGestureSuppressedTarget,
  shouldTriggerBack,
  shouldTriggerPull,
} from "./mobile-gestures";

describe("mobile gestures", () => {
  it("detects back-edge starts", () => {
    expect(isBackEdgeStart(0)).toBe(true);
    expect(isBackEdgeStart(28)).toBe(true);
    expect(isBackEdgeStart(29)).toBe(false);
  });

  it("clamps pull and back offsets", () => {
    expect(clampPullOffset(0)).toBe(0);
    expect(clampPullOffset(40)).toBe(20);
    expect(clampPullOffset(400)).toBe(120);
    expect(clampBackOffset(0)).toBe(0);
    expect(clampBackOffset(100)).toBe(35);
    expect(clampBackOffset(400)).toBe(120);
  });

  it("requires pull distance before refresh", () => {
    expect(shouldTriggerPull(40)).toBe(false);
    expect(shouldTriggerPull(72)).toBe(true);
  });

  it("requires mostly horizontal back swipes", () => {
    expect(shouldTriggerBack(90, 10)).toBe(true);
    expect(shouldTriggerBack(90, 80)).toBe(false);
    expect(shouldTriggerBack(40, 0)).toBe(false);
  });

  it("ignores interactive targets for gesture capture", () => {
    const button = document.createElement("button");
    const overlay = document.createElement("div");
    overlay.className = "toc-overlay";
    const panel = document.createElement("div");
    overlay.appendChild(panel);
    panel.appendChild(button);
    document.body.appendChild(overlay);

    expect(isGestureSuppressedTarget(button)).toBe(true);
    expect(isGestureSuppressedTarget(overlay)).toBe(true);

    overlay.remove();
  });
});

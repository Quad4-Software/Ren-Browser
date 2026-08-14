// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  clampBackOffset,
  clampForwardOffset,
  clampPullOffset,
  getEffectiveScrollTop,
  isBackEdgeStart,
  isForwardEdgeStart,
  isGestureSuppressedTarget,
  shouldTriggerBack,
  shouldTriggerForward,
  shouldTriggerPull,
} from "./mobile-gestures";

describe("mobile gestures", () => {
  it("detects back-edge starts", () => {
    expect(isBackEdgeStart(0)).toBe(true);
    expect(isBackEdgeStart(28)).toBe(true);
    expect(isBackEdgeStart(29)).toBe(false);
  });

  it("detects forward-edge starts", () => {
    expect(isForwardEdgeStart(390, 390)).toBe(true);
    expect(isForwardEdgeStart(362, 390)).toBe(true);
    expect(isForwardEdgeStart(361, 390)).toBe(false);
    expect(isForwardEdgeStart(28, 0)).toBe(false);
  });

  it("clamps pull and back offsets", () => {
    expect(clampPullOffset(0)).toBe(0);
    expect(clampPullOffset(40)).toBe(20);
    expect(clampPullOffset(400)).toBe(120);
    expect(clampBackOffset(0)).toBe(0);
    expect(clampBackOffset(100)).toBe(35);
    expect(clampBackOffset(400)).toBe(120);
    expect(clampForwardOffset(0)).toBe(0);
    expect(clampForwardOffset(-100)).toBe(35);
    expect(clampForwardOffset(-400)).toBe(120);
    expect(clampForwardOffset(40)).toBe(0);
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

  it("requires mostly horizontal forward swipes", () => {
    expect(shouldTriggerForward(-90, 10)).toBe(true);
    expect(shouldTriggerForward(-90, 80)).toBe(false);
    expect(shouldTriggerForward(-40, 0)).toBe(false);
    expect(shouldTriggerForward(90, 0)).toBe(false);
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

  it("reads scroll position from scrollable ancestors", () => {
    const boundary = document.createElement("section");
    const parent = document.createElement("div");
    const child = document.createElement("div");
    parent.appendChild(child);
    boundary.appendChild(parent);
    document.body.appendChild(boundary);

    Object.defineProperty(parent, "scrollTop", { value: 24, configurable: true });

    expect(getEffectiveScrollTop(child, boundary)).toBe(24);
    expect(getEffectiveScrollTop(null, boundary)).toBe(0);

    boundary.remove();
  });
});

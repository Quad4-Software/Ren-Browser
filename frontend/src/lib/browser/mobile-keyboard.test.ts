// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  KEYBOARD_OPEN_INSET_MIN,
  MOBILE_INPUT_FONT_PX,
  isEditableKeyboardTarget,
  keyboardChromeState,
  parseNativeKeyboardEvent,
  visualViewportKeyboardInset,
} from "./mobile-keyboard";

describe("mobile keyboard chrome", () => {
  it("pins text inputs at 16px so WebView does not zoom on focus", () => {
    expect(MOBILE_INPUT_FONT_PX).toBe(16);
  });

  it("treats iOS visualViewport shrink as keyboard inset", () => {
    expect(visualViewportKeyboardInset(800, { height: 500, offsetTop: 0 })).toBe(300);
    expect(visualViewportKeyboardInset(800, { height: 800, offsetTop: 0 })).toBe(0);
    expect(visualViewportKeyboardInset(800, null)).toBe(0);
  });

  it("uses offsetTop when the visual viewport is panned", () => {
    expect(visualViewportKeyboardInset(800, { height: 500, offsetTop: 300 })).toBe(0);
    expect(visualViewportKeyboardInset(800, { height: 400, offsetTop: 200 })).toBe(200);
  });

  it("parses Wails common:keyboard payloads", () => {
    expect(parseNativeKeyboardEvent({ data: '{"visible":true,"height":320}' })).toEqual({
      visible: true,
      height: 320,
    });
    expect(parseNativeKeyboardEvent({ visible: false, height: 0 })).toEqual({
      visible: false,
      height: 0,
    });
    expect(parseNativeKeyboardEvent("not-json")).toEqual({ visible: false, height: 0 });
  });

  it("detects text fields that open the IME", () => {
    const text = document.createElement("input");
    text.type = "text";
    const url = document.createElement("input");
    url.type = "url";
    const checkbox = document.createElement("input");
    checkbox.type = "checkbox";
    const area = document.createElement("textarea");
    expect(isEditableKeyboardTarget(text)).toBe(true);
    expect(isEditableKeyboardTarget(url)).toBe(true);
    expect(isEditableKeyboardTarget(area)).toBe(true);
    expect(isEditableKeyboardTarget(checkbox)).toBe(false);
    expect(isEditableKeyboardTarget(document.createElement("button"))).toBe(false);
  });

  it("hides chrome when Android reports the IME without a visualViewport inset", () => {
    const state = keyboardChromeState({
      visualInset: 0,
      native: { visible: true, height: 280 },
      focusedEditable: false,
    });
    expect(state.imeInset).toBe(0);
    expect(state.keyboardOpen).toBe(true);
  });

  it("pads only the visual inset so adjustResize is not double counted", () => {
    const state = keyboardChromeState({
      visualInset: 240,
      native: { visible: true, height: 240 },
      focusedEditable: true,
    });
    expect(state.imeInset).toBe(240);
    expect(state.keyboardOpen).toBe(true);
  });

  it("opens chrome from a large visual inset without a native event", () => {
    const state = keyboardChromeState({
      visualInset: KEYBOARD_OPEN_INSET_MIN,
      native: { visible: false, height: 0 },
      focusedEditable: false,
    });
    expect(state.keyboardOpen).toBe(true);
  });
});

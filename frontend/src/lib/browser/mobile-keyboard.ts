// SPDX-License-Identifier: MIT

export const MOBILE_INPUT_FONT_PX = 16;
export const KEYBOARD_OPEN_INSET_MIN = 80;

export type VisualViewportLike = {
  height: number;
  offsetTop: number;
};

export type NativeKeyboardState = {
  visible: boolean;
  height: number;
};

export type KeyboardChromeState = {
  imeInset: number;
  keyboardOpen: boolean;
};

const NON_TEXT_INPUT_TYPES = new Set([
  "button",
  "checkbox",
  "radio",
  "file",
  "hidden",
  "image",
  "reset",
  "submit",
  "range",
  "color",
]);

export function visualViewportKeyboardInset(
  layoutHeight: number,
  viewport: VisualViewportLike | null | undefined,
): number {
  if (!viewport || layoutHeight <= 0) {
    return 0;
  }
  const visibleBottom = viewport.offsetTop + viewport.height;
  return Math.max(0, Math.round(layoutHeight - visibleBottom));
}

export function parseNativeKeyboardEvent(payload: unknown): NativeKeyboardState {
  let data: unknown = payload;
  if (data && typeof data === "object" && "data" in data) {
    data = (data as { data: unknown }).data;
  }
  if (typeof data === "string") {
    try {
      data = JSON.parse(data) as unknown;
    } catch {
      return { visible: false, height: 0 };
    }
  }
  if (!data || typeof data !== "object") {
    return { visible: false, height: 0 };
  }
  const rec = data as { visible?: unknown; height?: unknown };
  const height = Math.max(0, Number(rec.height) || 0);
  return { visible: rec.visible === true || height >= KEYBOARD_OPEN_INSET_MIN, height };
}

export function isEditableKeyboardTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) {
    return false;
  }
  if (target.isContentEditable) {
    return true;
  }
  const tag = target.tagName;
  if (tag === "TEXTAREA" || tag === "SELECT") {
    return true;
  }
  if (tag !== "INPUT") {
    return false;
  }
  const type = (target as HTMLInputElement).type || "text";
  return !NON_TEXT_INPUT_TYPES.has(type.toLowerCase());
}

export function keyboardChromeState(input: {
  visualInset: number;
  native: NativeKeyboardState;
  focusedEditable: boolean;
}): KeyboardChromeState {
  const keyboardOpen =
    input.focusedEditable || input.native.visible || input.visualInset >= KEYBOARD_OPEN_INSET_MIN;
  return { imeInset: Math.max(0, input.visualInset), keyboardOpen };
}

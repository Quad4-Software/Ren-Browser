import { expect } from "vitest";

function overflowBlocksHorizontal(style: CSSStyleDeclaration): boolean {
  if (/hidden|clip/.test(style.overflowX)) {
    return true;
  }
  return /hidden|clip/.test(style.overflow);
}

export function expectsEllipsis(el: Element, options?: { requireMinWidthZero?: boolean }): void {
  const style = getComputedStyle(el);
  expect(style.overflow).toMatch(/hidden|clip/);
  expect(style.textOverflow).toBe("ellipsis");
  expect(style.whiteSpace).toBe("nowrap");
  if (options?.requireMinWidthZero) {
    expect(["0px", "0"]).toContain(style.minWidth);
  }
}

export function expectsVerticalScroll(el: Element): void {
  const style = getComputedStyle(el);
  if (/auto|scroll/.test(style.overflowY)) {
    return;
  }
  expect(style.overflow).toMatch(/auto|scroll/);
}

export function expectsNoHorizontalScroll(el: Element): void {
  expect(overflowBlocksHorizontal(getComputedStyle(el))).toBe(true);
}

export function parsePx(value: string): number {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

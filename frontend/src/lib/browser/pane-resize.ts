// SPDX-License-Identifier: MIT

export const SIDEBAR_MIN_PX = 280;
export const SIDEBAR_MAX_PX = 560;
export const SIDEBAR_DEFAULT_PX = 360;
export const SPLIT_RATIO_MIN = 25;
export const SPLIT_RATIO_MAX = 75;

export function clampNumber(value: number, min: number, max: number): number {
  if (!Number.isFinite(value)) {
    return min;
  }
  return Math.min(max, Math.max(min, value));
}

export function clampSplitRatio(
  ratio: number,
  min = SPLIT_RATIO_MIN,
  max = SPLIT_RATIO_MAX,
): number {
  return clampNumber(ratio, min, max);
}

export function clampSidebarWidth(width: number, containerWidth: number): number {
  const usable = Number.isFinite(containerWidth) ? containerWidth : SIDEBAR_MAX_PX;
  const max = Math.min(SIDEBAR_MAX_PX, Math.max(SIDEBAR_MIN_PX, Math.floor(usable * 0.5)));
  return Math.round(clampNumber(width, SIDEBAR_MIN_PX, max));
}

export function splitRatioFromPointer(
  event: PointerEvent,
  rect: DOMRect,
  vertical = false,
): number {
  const span = vertical ? rect.height : rect.width;
  if (span <= 0) {
    return 50;
  }
  const next = vertical
    ? ((event.clientY - rect.top) / span) * 100
    : ((event.clientX - rect.left) / span) * 100;
  return clampSplitRatio(next);
}

export function sidebarWidthFromPointer(event: PointerEvent, rect: DOMRect): number {
  return clampSidebarWidth(rect.right - event.clientX, rect.width);
}

export function capturePointer(target: HTMLElement, event: PointerEvent): void {
  if (target.hasPointerCapture(event.pointerId)) {
    return;
  }
  target.setPointerCapture(event.pointerId);
}

export function releasePointer(target: HTMLElement | null, event: PointerEvent): void {
  if (target?.hasPointerCapture(event.pointerId)) {
    target.releasePointerCapture(event.pointerId);
  }
}

export function createDragScheduler(apply: (event: PointerEvent) => void): {
  move: (event: PointerEvent) => void;
  cancel: () => void;
} {
  let frame = 0;
  let pending: PointerEvent | null = null;

  function move(event: PointerEvent) {
    pending = event;
    if (frame) {
      return;
    }
    frame = requestAnimationFrame(() => {
      frame = 0;
      const next = pending;
      pending = null;
      if (next) {
        apply(next);
      }
    });
  }

  function cancel() {
    if (frame) {
      cancelAnimationFrame(frame);
      frame = 0;
    }
    pending = null;
  }

  return { move, cancel };
}

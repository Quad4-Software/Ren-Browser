// SPDX-License-Identifier: MIT

export const MOBILE_BACK_EDGE_WIDTH = 28;
export const MOBILE_PULL_THRESHOLD = 72;
export const MOBILE_BACK_THRESHOLD = 80;
export const MOBILE_GESTURE_MAX_VERTICAL_DRIFT = 56;
export const MOBILE_PULL_MAX_OFFSET = 120;
export const MOBILE_BACK_MAX_OFFSET = 120;

export type MobileGestureProgress = {
  pullOffset: number;
  pullTriggered: boolean;
  backOffset: number;
  backTriggered: boolean;
};

export type MobileGestureOptions = {
  getCanGoBack: () => boolean;
  getScrollTop: () => number;
  isActive: () => boolean;
  onRefresh: () => void;
  onBack: () => void;
  onProgress?: (progress: MobileGestureProgress) => void;
};

export const GESTURE_SUPPRESSED_SELECTOR =
  'button, a, input, textarea, select, label, summary, [role="button"], [role="dialog"], [role="menu"], [role="menuitem"], [contenteditable="true"], .toc-overlay, .reader-search';

export function isGestureSuppressedTarget(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) {
    return false;
  }
  return Boolean(target.closest(GESTURE_SUPPRESSED_SELECTOR));
}

export function isBackEdgeStart(clientX: number, edgeWidth = MOBILE_BACK_EDGE_WIDTH): boolean {
  return clientX <= edgeWidth;
}

export function shouldTriggerPull(deltaY: number, threshold = MOBILE_PULL_THRESHOLD): boolean {
  return deltaY >= threshold;
}

export function shouldTriggerBack(
  deltaX: number,
  deltaY: number,
  threshold = MOBILE_BACK_THRESHOLD,
): boolean {
  return deltaX >= threshold && Math.abs(deltaY) <= MOBILE_GESTURE_MAX_VERTICAL_DRIFT;
}

export function clampPullOffset(deltaY: number): number {
  if (deltaY <= 0) {
    return 0;
  }
  return Math.min(deltaY * 0.5, MOBILE_PULL_MAX_OFFSET);
}

export function clampBackOffset(deltaX: number): number {
  if (deltaX <= 0) {
    return 0;
  }
  return Math.min(deltaX * 0.35, MOBILE_BACK_MAX_OFFSET);
}

function emptyProgress(): MobileGestureProgress {
  return {
    pullOffset: 0,
    pullTriggered: false,
    backOffset: 0,
    backTriggered: false,
  };
}

export function getEffectiveScrollTop(
  element: HTMLElement | null | undefined,
  boundary?: HTMLElement | null,
): number {
  if (!element) {
    return 0;
  }

  let current: HTMLElement | null = element;
  while (current) {
    if (current.scrollTop > 1) {
      return current.scrollTop;
    }
    if (boundary && current === boundary) {
      break;
    }
    current = current.parentElement;
  }

  return 0;
}

export function attachMobileGestures(
  surface: HTMLElement,
  options: MobileGestureOptions,
): { teardown: () => void } {
  let activePointerId: number | null = null;
  let mode: "none" | "pull" | "back" = "none";
  let startX = 0;
  let startY = 0;

  const report = (progress: MobileGestureProgress) => {
    options.onProgress?.(progress);
  };

  const resetProgress = () => {
    report(emptyProgress());
  };

  const finish = (event: PointerEvent) => {
    if (surface.hasPointerCapture(event.pointerId)) {
      surface.releasePointerCapture(event.pointerId);
    }
    activePointerId = null;
    mode = "none";
    resetProgress();
  };

  const onPointerDown = (event: PointerEvent) => {
    if (!options.isActive() || activePointerId !== null) {
      return;
    }
    if (event.pointerType === "mouse" && event.button !== 0) {
      return;
    }
    if (isGestureSuppressedTarget(event.target)) {
      return;
    }

    const scrollTop = options.getScrollTop();
    if (options.getCanGoBack() && isBackEdgeStart(event.clientX)) {
      mode = "back";
      activePointerId = event.pointerId;
      startX = event.clientX;
      startY = event.clientY;
      surface.setPointerCapture(event.pointerId);
      return;
    }

    if (scrollTop <= 1) {
      mode = "pull";
      activePointerId = event.pointerId;
      startX = event.clientX;
      startY = event.clientY;
      surface.setPointerCapture(event.pointerId);
    }
  };

  const onPointerMove = (event: PointerEvent) => {
    if (activePointerId !== event.pointerId || mode === "none") {
      return;
    }

    const deltaX = event.clientX - startX;
    const deltaY = event.clientY - startY;

    if (mode === "back") {
      if (deltaX < 0) {
        finish(event);
        return;
      }
      if (Math.abs(deltaY) > MOBILE_GESTURE_MAX_VERTICAL_DRIFT && deltaX < 20) {
        finish(event);
        return;
      }
      report({
        pullOffset: 0,
        pullTriggered: false,
        backOffset: clampBackOffset(deltaX),
        backTriggered: shouldTriggerBack(deltaX, deltaY),
      });
      event.preventDefault();
      return;
    }

    if (mode === "pull") {
      if (deltaY < 0 || options.getScrollTop() > 1) {
        finish(event);
        return;
      }
      report({
        pullOffset: clampPullOffset(deltaY),
        pullTriggered: shouldTriggerPull(deltaY),
        backOffset: 0,
        backTriggered: false,
      });
      event.preventDefault();
    }
  };

  const onPointerUp = (event: PointerEvent) => {
    if (activePointerId !== event.pointerId) {
      return;
    }

    const deltaX = event.clientX - startX;
    const deltaY = event.clientY - startY;

    if (mode === "back" && options.getCanGoBack() && shouldTriggerBack(deltaX, deltaY)) {
      options.onBack();
    } else if (mode === "pull" && shouldTriggerPull(deltaY)) {
      options.onRefresh();
    }

    finish(event);
  };

  surface.addEventListener("pointerdown", onPointerDown);
  surface.addEventListener("pointermove", onPointerMove);
  surface.addEventListener("pointerup", onPointerUp);
  surface.addEventListener("pointercancel", onPointerUp);

  return {
    teardown: () => {
      surface.removeEventListener("pointerdown", onPointerDown);
      surface.removeEventListener("pointermove", onPointerMove);
      surface.removeEventListener("pointerup", onPointerUp);
      surface.removeEventListener("pointercancel", onPointerUp);
      resetProgress();
    },
  };
}

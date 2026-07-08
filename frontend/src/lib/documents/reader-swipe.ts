// SPDX-License-Identifier: MIT
import { READER_SWIPE_MAX_VERTICAL, READER_SWIPE_THRESHOLD } from "./reader-layout";

export type ReaderSwipeOptions = {
  isEnabled: () => boolean;
  canPrev: () => boolean;
  canNext: () => boolean;
  onPrev: () => void;
  onNext: () => void;
};

export function attachReaderSwipe(
  surface: HTMLElement,
  options: ReaderSwipeOptions,
): { teardown: () => void } {
  let activePointerId: number | null = null;
  let startX = 0;
  let startY = 0;

  const finish = (event: PointerEvent) => {
    if (surface.hasPointerCapture(event.pointerId)) {
      surface.releasePointerCapture(event.pointerId);
    }
    activePointerId = null;
  };

  const onPointerDown = (event: PointerEvent) => {
    if (!options.isEnabled() || activePointerId !== null) {
      return;
    }
    if (event.pointerType === "mouse" && event.button !== 0) {
      return;
    }
    if (
      event.target instanceof Element &&
      event.target.closest("button, a, input, textarea, select, label")
    ) {
      return;
    }
    activePointerId = event.pointerId;
    startX = event.clientX;
    startY = event.clientY;
    surface.setPointerCapture(event.pointerId);
  };

  const onPointerMove = (event: PointerEvent) => {
    if (activePointerId !== event.pointerId) {
      return;
    }
    const deltaX = event.clientX - startX;
    const deltaY = event.clientY - startY;
    if (Math.abs(deltaX) > 12 && Math.abs(deltaY) > READER_SWIPE_MAX_VERTICAL) {
      finish(event);
      return;
    }
    if (Math.abs(deltaX) > 12 && Math.abs(deltaY) <= READER_SWIPE_MAX_VERTICAL) {
      event.preventDefault();
    }
  };

  const onPointerUp = (event: PointerEvent) => {
    if (activePointerId !== event.pointerId) {
      return;
    }
    const deltaX = event.clientX - startX;
    const deltaY = event.clientY - startY;
    if (
      Math.abs(deltaX) >= READER_SWIPE_THRESHOLD &&
      Math.abs(deltaY) <= READER_SWIPE_MAX_VERTICAL
    ) {
      if (deltaX < 0 && options.canNext()) {
        options.onNext();
      } else if (deltaX > 0 && options.canPrev()) {
        options.onPrev();
      }
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
    },
  };
}

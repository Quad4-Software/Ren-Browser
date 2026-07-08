// SPDX-License-Identifier: MIT

export function isTypingTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) {
    return false;
  }
  const tag = target.tagName;
  return tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT" || target.isContentEditable;
}

export function handleDocumentArrowKeys(
  event: KeyboardEvent,
  opts: {
    viewport?: HTMLElement | null;
    onPrev: () => void;
    onNext: () => void;
    canPrev: boolean;
    canNext: boolean;
  },
): void {
  if (isTypingTarget(event.target) || event.altKey || event.ctrlKey || event.metaKey) {
    return;
  }
  if (event.key === "ArrowLeft" && opts.canPrev) {
    event.preventDefault();
    opts.onPrev();
    return;
  }
  if (event.key === "ArrowRight" && opts.canNext) {
    event.preventDefault();
    opts.onNext();
    return;
  }
  if ((event.key === "ArrowUp" || event.key === "ArrowDown") && opts.viewport) {
    event.preventDefault();
    opts.viewport.scrollBy({ top: event.key === "ArrowUp" ? -72 : 72, behavior: "smooth" });
  }
}

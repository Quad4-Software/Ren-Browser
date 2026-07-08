// SPDX-License-Identifier: MIT

export function frameSearchRoot(frame: HTMLIFrameElement | null | undefined): HTMLElement | null {
  const doc = frame?.contentDocument;
  if (!doc) {
    return null;
  }
  const root = doc.querySelector(".reader-root");
  return root instanceof HTMLElement ? root : doc.body;
}

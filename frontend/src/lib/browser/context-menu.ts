// SPDX-License-Identifier: MIT

export function clampMenuPosition(
  x: number,
  y: number,
  width: number,
  height: number,
  margin = 8,
): { x: number; y: number } {
  const viewportW = typeof window !== "undefined" ? window.innerWidth : width + margin * 2;
  const viewportH = typeof window !== "undefined" ? window.innerHeight : height + margin * 2;
  const maxX = Math.max(margin, viewportW - width - margin);
  const maxY = Math.max(margin, viewportH - height - margin);
  return {
    x: Math.min(Math.max(x, margin), maxX),
    y: Math.min(Math.max(y, margin), maxY),
  };
}

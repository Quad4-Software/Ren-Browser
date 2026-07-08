// SPDX-License-Identifier: MIT

export type ReaderRotation = 0 | 90 | 180 | 270;

export const READER_FONT_SCALES = [0.85, 1, 1.15, 1.3, 1.5] as const;

export const READER_SWIPE_THRESHOLD = 64;
export const READER_SWIPE_MAX_VERTICAL = 48;

export function cycleReaderRotation(current: ReaderRotation): ReaderRotation {
  switch (current) {
    case 0:
      return 90;
    case 90:
      return 180;
    case 180:
      return 270;
    default:
      return 0;
  }
}

export function stepReaderFontScale(current: number, direction: 1 | -1): number {
  const scales = READER_FONT_SCALES;
  let index = scales.findIndex((scale) => Math.abs(scale - current) < 0.001);
  if (index < 0) {
    index = scales.findIndex((scale) => scale >= current);
    if (index < 0) {
      index = scales.length - 1;
    }
  }
  const next = Math.min(scales.length - 1, Math.max(0, index + direction));
  return scales[next] ?? 1;
}

export function readerFontScaleLabel(scale: number): string {
  return `${Math.round(scale * 100)}%`;
}

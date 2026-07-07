// SPDX-License-Identifier: MIT

export type MicronSegment = { type: "text" | "markup"; value: string };

export function splitMicronSegments(source: string): MicronSegment[] {
  const segments: MicronSegment[] = [];
  let i = 0;
  while (i < source.length) {
    if (source[i] === "`") {
      let j = i + 1;
      while (j < source.length && source[j] !== "`") {
        j += 1;
      }
      if (j < source.length) {
        segments.push({ type: "markup", value: source.slice(i, j + 1) });
        i = j + 1;
        continue;
      }
    }
    let j = i;
    while (j < source.length && source[j] !== "`") {
      j += 1;
    }
    const chunk = source.slice(i, j);
    if (chunk) {
      segments.push({ type: "text", value: chunk });
    }
    i = j;
  }
  return segments;
}

export function joinMicronSegments(segments: MicronSegment[]): string {
  return segments.map((segment) => segment.value).join("");
}

export function isTranslatableText(text: string): boolean {
  return /[\p{L}\p{N}]/u.test(text.trim());
}

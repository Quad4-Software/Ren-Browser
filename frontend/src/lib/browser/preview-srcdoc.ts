// SPDX-License-Identifier: MIT

export const PREVIEW_REF_WIDTH = 1280;
export const PREVIEW_REF_HEIGHT = 1000;
const PREVIEW_FALLBACK_SCALE = 0.22;

export function previewScaleForBox(boxWidth: number, refWidth = PREVIEW_REF_WIDTH): number {
  if (boxWidth <= 0) {
    return PREVIEW_FALLBACK_SCALE;
  }
  return boxWidth / refWidth;
}

export function wrapPreviewSrcdoc(html: string, colors?: { fg?: string; bg?: string }): string {
  const trimmed = html.trim();
  if (!trimmed) {
    return "";
  }
  if (/^<!DOCTYPE/i.test(trimmed) || /^<html[\s>]/i.test(trimmed)) {
    return trimmed;
  }
  const fg = colors?.fg?.trim() || "#ffffff";
  const bg = colors?.bg?.trim() || "#000000";
  return `<!DOCTYPE html><html><head><meta charset="utf-8"><style>
html,body{margin:0;padding:0;width:${PREVIEW_REF_WIDTH}px;min-height:${PREVIEW_REF_HEIGHT}px;overflow:hidden;background:${bg};color:${fg};}
body{padding:12px;box-sizing:border-box;font:14px/1.45 system-ui,sans-serif;}
img,video,svg,table{max-width:100%;height:auto;}
pre{white-space:pre-wrap;word-break:break-word;max-width:100%;margin:0;font:inherit;}
</style></head><body>${trimmed}</body></html>`;
}

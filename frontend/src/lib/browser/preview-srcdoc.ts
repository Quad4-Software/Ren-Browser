// SPDX-License-Identifier: MIT

export const PREVIEW_REF_WIDTH = 1280;
export const PREVIEW_REF_HEIGHT = 1000;
const PREVIEW_FALLBACK_SCALE = 0.22;
const SAFE_CSS_COLOR_RE = /^#[0-9a-fA-F]{3}$|^#[0-9a-fA-F]{6}$|^#[0-9a-fA-F]{8}$/;

export function previewScaleForBox(boxWidth: number, refWidth = PREVIEW_REF_WIDTH): number {
  if (boxWidth <= 0) {
    return PREVIEW_FALLBACK_SCALE;
  }
  return boxWidth / refWidth;
}

function safePreviewColor(value: string | undefined, fallback: string): string {
  const trimmed = value?.trim() ?? "";
  return SAFE_CSS_COLOR_RE.test(trimmed) ? trimmed : fallback;
}

function stripPreviewExecutableMarkup(html: string): string {
  return html
    .replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, "")
    .replace(/<script\b[^>]*\/>/gi, "")
    .replace(/\son[a-z]+\s*=\s*(['"]).*?\1/gi, "")
    .replace(/\son[a-z]+\s*=\s*[^\s>]+/gi, "")
    .replace(/javascript\s*:/gi, "blocked:");
}

export function wrapPreviewSrcdoc(html: string, colors?: { fg?: string; bg?: string }): string {
  const trimmed = html.trim();
  if (!trimmed) {
    return "";
  }
  if (/^<!DOCTYPE/i.test(trimmed) || /^<html[\s>]/i.test(trimmed)) {
    return stripPreviewExecutableMarkup(trimmed);
  }
  const fg = safePreviewColor(colors?.fg, "#ffffff");
  const bg = safePreviewColor(colors?.bg, "#000000");
  return `<!DOCTYPE html><html><head><meta charset="utf-8"><style>
html,body{margin:0;padding:0;width:${PREVIEW_REF_WIDTH}px;min-height:${PREVIEW_REF_HEIGHT}px;overflow:hidden;background:${bg};color:${fg};}
body{padding:12px;box-sizing:border-box;font:14px/1.45 system-ui,sans-serif;}
img,video,svg,table{max-width:100%;height:auto;}
pre{white-space:pre-wrap;word-break:break-word;max-width:100%;margin:0;font:inherit;}
</style></head><body>${trimmed}</body></html>`;
}

// SPDX-License-Identifier: MIT
import { readerColors, type ReaderTheme } from "$lib/documents/reader-theme";

export type ReaderDisplayOptions = {
  theme: ReaderTheme;
  fontScale?: number;
  rotation?: number;
};

const DOCUMENT_FRAME_CSP =
  "default-src 'none'; base-uri 'none'; form-action 'none'; connect-src 'none'; font-src 'none'; media-src 'none'; object-src 'none'; frame-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src blob: data: image/*";

function readerShellStyle(options: ReaderDisplayOptions): string {
  const colors = readerColors(options.theme);
  const fontScale = options.fontScale ?? 1;
  const rotation = options.rotation ?? 0;
  return `
html, body {
  margin: 0;
  padding: 0;
  background: ${colors.bg};
  color: ${colors.fg};
  line-height: 1.6;
  overflow-wrap: anywhere;
  height: 100%;
  overflow: auto;
  color-scheme: ${options.theme};
}
.reader-stage {
  min-height: 100%;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 0.5rem 0 1.5rem;
}
.reader-root {
  width: 100%;
  max-width: 42rem;
  font-size: ${fontScale}rem;
  transform: rotate(${rotation}deg);
  transform-origin: center top;
}
.reader-root *:not(img):not(svg):not(image) {
  background: transparent !important;
  background-color: transparent !important;
  color: inherit !important;
  border-color: ${colors.border} !important;
}
img {
  max-width: 100%;
  height: auto;
}
a {
  pointer-events: none;
  color: inherit;
  text-decoration: underline;
}
`;
}

export function buildIsolatedHtmlDocument(bodyHtml: string, options: ReaderDisplayOptions): string {
  return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="${DOCUMENT_FRAME_CSP}">
<style>${readerShellStyle(options)}</style>
</head>
<body><div class="reader-stage"><div class="reader-root">${bodyHtml}</div></div></body>
</html>`;
}

export const ISOLATED_FRAME_SANDBOX = "allow-same-origin";

const PDF_FRAME_CSP =
  "default-src 'none'; base-uri 'none'; form-action 'none'; connect-src 'none'; font-src 'none'; media-src 'none'; object-src 'none'; frame-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src 'none'";

function pdfShellStyle(theme: ReaderTheme, rotation: number): string {
  const colors = readerColors(theme);
  return `
html, body {
  margin: 0;
  padding: 0;
  background: ${colors.bg};
  height: 100%;
  overflow: auto;
  color-scheme: ${theme};
}
body {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 1rem 0 2rem;
}
.page-stage {
  transform: rotate(${rotation}deg);
  transform-origin: center center;
}
.page-canvas {
  max-width: 100%;
  height: auto;
  box-shadow: 0 2px 12px rgb(0 0 0 / ${theme === "dark" ? "0.45" : "0.12"});
  background: #fff;
  display: block;
}
`;
}

export function buildIsolatedPdfShell(theme: ReaderTheme, rotation = 0): string {
  return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="Content-Security-Policy" content="${PDF_FRAME_CSP}">
<style>${pdfShellStyle(theme, rotation)}</style>
</head>
<body><div class="page-stage"><canvas id="page-canvas" class="page-canvas"></canvas></div></body>
</html>`;
}

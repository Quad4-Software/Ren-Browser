// SPDX-License-Identifier: MIT
import * as pdfjs from "pdfjs-dist";
import pdfWorker from "pdfjs-dist/build/pdf.worker.min.mjs?url";
import { DOCUMENT_PDF_RENDER_TIMEOUT_MS, withTimeout } from "./async";

pdfjs.GlobalWorkerOptions.workerSrc = pdfWorker;

export type PdfLoadResult = {
  doc: pdfjs.PDFDocumentProxy;
  numPages: number;
};

export async function loadPdfDocument(data: Uint8Array): Promise<PdfLoadResult> {
  const copy = new Uint8Array(data);
  const loadingTask = pdfjs.getDocument({
    data: copy,
    useWorkerFetch: false,
    useSystemFonts: false,
    disableFontFace: true,
    disableAutoFetch: true,
    disableStream: true,
    disableRange: true,
  });
  const doc = await withTimeout(loadingTask.promise, DOCUMENT_PDF_RENDER_TIMEOUT_MS, "PDF load");
  return { doc, numPages: doc.numPages };
}

export async function renderPdfPage(
  doc: pdfjs.PDFDocumentProxy,
  pageNumber: number,
  canvas: HTMLCanvasElement,
  scale: number,
  rotation = 0,
): Promise<void> {
  const page = await withTimeout(
    doc.getPage(pageNumber),
    DOCUMENT_PDF_RENDER_TIMEOUT_MS,
    "PDF page load",
  );
  const viewport = page.getViewport({ scale, rotation });
  const context = canvas.getContext("2d");
  if (!context) {
    throw new Error("canvas context unavailable");
  }
  canvas.width = Math.floor(viewport.width);
  canvas.height = Math.floor(viewport.height);
  await withTimeout(
    page.render({ canvasContext: context, viewport, canvas }).promise,
    DOCUMENT_PDF_RENDER_TIMEOUT_MS,
    "PDF page render",
  );
}

export function fitWidthScale(page: pdfjs.PDFPageProxy, containerWidth: number): number {
  const viewport = page.getViewport({ scale: 1 });
  if (viewport.width <= 0 || containerWidth <= 0) {
    return 1;
  }
  return containerWidth / viewport.width;
}

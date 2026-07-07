// SPDX-License-Identifier: MIT

export class DocumentTimeoutError extends Error {
  constructor(label: string, ms: number) {
    super(`${label} timed out after ${Math.round(ms / 1000)}s`);
    this.name = "DocumentTimeoutError";
  }
}

export function withTimeout<T>(promise: Promise<T>, ms: number, label: string): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new DocumentTimeoutError(label, ms));
    }, ms);
    promise.then(
      (value) => {
        clearTimeout(timer);
        resolve(value);
      },
      (err) => {
        clearTimeout(timer);
        reject(err);
      },
    );
  });
}

export const DOCUMENT_PARSE_TIMEOUT_MS = 120_000;
export const DOCUMENT_PDF_RENDER_TIMEOUT_MS = 60_000;

export function formatDocumentError(err: unknown, fallback: string): string {
  if (err instanceof DocumentTimeoutError) {
    return err.message;
  }
  if (err instanceof DOMException && err.name === "InvalidCharacterError") {
    return "document data is corrupted or incomplete";
  }
  if (err instanceof Error && err.message.trim()) {
    return err.message;
  }
  return fallback;
}

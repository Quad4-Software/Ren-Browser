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

export function documentErrorI18nKey(err: unknown): string | null {
  const message =
    err instanceof Error ? err.message.toLowerCase() : String(err ?? "").toLowerCase();
  if (
    message.includes("end of central directory") ||
    message.includes("corrupted zip") ||
    message.includes("incomplete or corrupted") ||
    message.includes("missing zip header") ||
    message.includes("too small or incomplete")
  ) {
    return "documents.corruptEpub";
  }
  return null;
}

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

export function resolveDocumentErrorMessage(
  err: unknown,
  translate: (key: string) => string,
): string {
  const key = documentErrorI18nKey(err);
  if (key) {
    return translate(key);
  }
  return formatDocumentError(err, translate("documents.loadFailed"));
}

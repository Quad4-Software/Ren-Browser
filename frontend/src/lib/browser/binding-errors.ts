// SPDX-License-Identifier: MIT

export type WailsRuntimeError = {
  message?: string;
  kind?: string;
  cause?: unknown;
};

export function errorText(err: unknown): string {
  if (err == null) {
    return "";
  }
  if (err instanceof Error) {
    return err.message;
  }
  if (typeof err === "string") {
    return err;
  }
  return String(err);
}

export function parseWailsRuntimeError(text: string): WailsRuntimeError | null {
  const trimmed = text.trim();
  if (!trimmed.startsWith("{")) {
    return null;
  }
  try {
    const parsed = JSON.parse(trimmed) as WailsRuntimeError;
    if (parsed && typeof parsed === "object" && typeof parsed.message === "string") {
      return parsed;
    }
  } catch {
    return null;
  }
  return null;
}

export function unwrapBindingErrorMessage(text: string): string {
  const trimmed = text.trim();
  if (!trimmed) {
    return "";
  }
  const wails = parseWailsRuntimeError(trimmed);
  if (wails?.message?.trim()) {
    return wails.message.trim();
  }
  return trimmed;
}

export function formatBindingError(
  err: unknown,
  fallback = "An unexpected error occurred",
): string {
  const message = unwrapBindingErrorMessage(errorText(err));
  return message || fallback;
}

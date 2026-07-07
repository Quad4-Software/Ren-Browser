// SPDX-License-Identifier: MIT

import { formatBindingError } from "$lib/browser/binding-errors.js";

export function crashErrorMessage(error: unknown): string {
  return formatBindingError(error, "Unknown error");
}

export function buildCrashDebugLog(message: string, cause?: unknown): string {
  const lines = [
    "Ren Browser crash report",
    `Time: ${new Date().toISOString()}`,
    `URL: ${typeof location === "undefined" ? "" : location.href}`,
    `User-Agent: ${typeof navigator === "undefined" ? "" : navigator.userAgent}`,
    "",
    `Error: ${message}`,
  ];
  if (cause instanceof Error) {
    if (cause.name && cause.name !== "Error") {
      lines.push(`Type: ${cause.name}`);
    }
    if (cause.stack?.trim()) {
      lines.push("", "Stack:", cause.stack.trim());
    }
  } else if (cause != null) {
    lines.push("", "Details:", crashErrorMessage(cause));
  }
  return `${lines.join("\n")}\n`;
}

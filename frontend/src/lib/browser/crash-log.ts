// SPDX-License-Identifier: MIT

export function crashErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message.trim() || error.name || "Error";
  }
  if (typeof error === "string") {
    return error.trim() || "Error";
  }
  if (error == null) {
    return "Unknown error";
  }
  try {
    return JSON.stringify(error);
  } catch {
    return String(error);
  }
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

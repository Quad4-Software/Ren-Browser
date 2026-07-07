// SPDX-License-Identifier: MIT

import { translate, translatePermission } from "$lib/i18n/catalog.js";

export type WailsRuntimeError = {
  message?: string;
  kind?: string;
  cause?: unknown;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return value != null && typeof value === "object";
}

export function isWailsRuntimeErrorPayload(value: unknown): value is WailsRuntimeError {
  if (!isRecord(value)) {
    return false;
  }
  return typeof value.message === "string";
}

export function errorText(err: unknown): string {
  if (err == null) {
    return "";
  }
  if (typeof err === "string") {
    return err;
  }
  if (isWailsRuntimeErrorPayload(err)) {
    return err.message?.trim() ?? "";
  }
  if (err instanceof Error) {
    return err.message;
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

function formatPluginId(pluginId: string): string {
  const short = pluginId.replace(/^renbrowser\./, "");
  if (!short) {
    return pluginId;
  }
  return short
    .split(/[.-]/g)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export function humanizeBindingError(message: string): string {
  const trimmed = message.trim();
  if (!trimmed) {
    return "";
  }

  let match = trimmed.match(/^plugin ([\w.-]+) permission ([\w.]+) not granted$/);
  if (match) {
    const [, pluginId, perm] = match;
    return translate("errors.extensionPermissionNotGranted", {
      plugin: formatPluginId(pluginId),
      permission: translatePermission(perm),
    });
  }

  match = trimmed.match(/^plugin ([\w.-]+) lacks permission ([\w.]+)$/);
  if (match) {
    const [, pluginId, perm] = match;
    return translate("errors.extensionPermissionLacks", {
      plugin: formatPluginId(pluginId),
      permission: translatePermission(perm),
    });
  }

  match = trimmed.match(/^extension exceeded network request limit \((\d+)\)$/i);
  if (match) {
    return translate("errors.extensionNetworkLimit", { limit: match[1] });
  }

  match = trimmed.match(/^plugin "([^"]+)" not found$/);
  if (match) {
    return translate("errors.extensionNotFound", { plugin: formatPluginId(match[1]) });
  }

  return trimmed;
}

export function formatBindingError(
  err: unknown,
  fallback = "An unexpected error occurred",
): string {
  const message = humanizeBindingError(unwrapBindingErrorMessage(errorText(err)));
  return message || fallback;
}

export function toBindingError(err: unknown, fallback = "An unexpected error occurred"): Error {
  return new Error(formatBindingError(err, fallback));
}

// SPDX-License-Identifier: MIT
import { ReportPluginError } from "../../../bindings/renbrowser/internal/app/pluginhost.js";
import { formatBindingError } from "$lib/browser/binding-errors.js";
import { deactivatePlugin } from "./lifecycle.js";

export async function reportPluginFailure(
  pluginId: string,
  phase: string,
  err: unknown,
): Promise<void> {
  const message = formatBindingError(err, "Extension failed");
  const detail = err instanceof Error ? (err.stack ?? "") : "";
  try {
    await deactivatePlugin(pluginId);
  } catch {
    // Best effort unload of the frontend module.
  }
  try {
    await ReportPluginError(pluginId, phase, message, detail);
  } catch {
    // Host may already have disabled the plugin.
  }
}

export { formatBindingError as formatPluginError } from "$lib/browser/binding-errors.js";

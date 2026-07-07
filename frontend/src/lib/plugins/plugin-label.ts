// SPDX-License-Identifier: MIT
import { getUILocale } from "$lib/i18n/i18n.svelte";
import { resolvePluginLabel } from "./plugin-i18n.js";

export function pluginLabel(pluginId: string, title: string): string {
  void getUILocale();
  return resolvePluginLabel(pluginId, title, getUILocale());
}

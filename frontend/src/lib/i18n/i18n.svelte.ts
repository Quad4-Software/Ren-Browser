// SPDX-License-Identifier: MIT
import { detectOSLocale, setCatalogLocale, translate, type TranslateParams } from "./catalog";
import { type LocaleCode, resolveLocale } from "./locales";

export let uiLocale = $state<LocaleCode>(resolveLocale(detectOSLocale()));

export function t(key: string, params?: TranslateParams): string {
  void uiLocale;
  return translate(key, params);
}

export function setUILocale(locale: string): LocaleCode {
  const resolved = setCatalogLocale(locale);
  uiLocale = resolved;
  return resolved;
}

export function initUILocale(saved?: string): LocaleCode {
  const initial = saved?.trim() ? resolveLocale(saved) : resolveLocale(detectOSLocale());
  return setUILocale(initial);
}

export { detectOSLocale } from "./catalog";
export {
  DEFAULT_LOCALE,
  SUPPORTED_LOCALES,
  localeLabel,
  resolveLocale,
  type LocaleCode,
} from "./locales";

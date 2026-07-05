// SPDX-License-Identifier: MIT
import { detectOSLocale, setCatalogLocale, translate, type TranslateParams } from "./catalog";
import { type LocaleCode, resolveLocale } from "./locales";

const locale = $state<{ code: LocaleCode }>({
  code: resolveLocale(detectOSLocale()),
});

export function getUILocale(): LocaleCode {
  return locale.code;
}

export function t(key: string, params?: TranslateParams): string {
  void locale.code;
  return translate(key, params);
}

export function setUILocale(localeTag: string): LocaleCode {
  const resolved = setCatalogLocale(localeTag);
  locale.code = resolved;
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
  localeNativeName,
  localeLabel,
  resolveLocale,
  type LocaleCode,
} from "./locales";

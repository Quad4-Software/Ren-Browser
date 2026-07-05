// SPDX-License-Identifier: MIT
export const DEFAULT_LOCALE = "en" as const;

export const SUPPORTED_LOCALES = [
  { code: "en", label: "English", nativeLabel: "English" },
  { code: "de", label: "German", nativeLabel: "Deutsch" },
  { code: "es", label: "Spanish", nativeLabel: "Español" },
  { code: "ru", label: "Russian", nativeLabel: "Русский" },
] as const;

export type LocaleCode = (typeof SUPPORTED_LOCALES)[number]["code"];

const localeCodes = new Set<string>(SUPPORTED_LOCALES.map((entry) => entry.code));

export function isLocaleCode(value: string): value is LocaleCode {
  return localeCodes.has(value);
}

export function resolveLocale(input: string): LocaleCode {
  const normalized = input.trim().toLowerCase().replaceAll("_", "-");
  if (!normalized) {
    return DEFAULT_LOCALE;
  }
  if (isLocaleCode(normalized)) {
    return normalized;
  }
  const prefix = normalized.split("-")[0];
  if (isLocaleCode(prefix)) {
    return prefix;
  }
  return DEFAULT_LOCALE;
}

export function localeLabel(code: string): string {
  const entry = SUPPORTED_LOCALES.find((item) => item.code === code);
  return entry ? `${entry.nativeLabel} (${entry.label})` : code;
}

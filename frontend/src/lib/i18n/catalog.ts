// SPDX-License-Identifier: MIT
import de from "./locales/de.json";
import en from "./locales/en.json";
import es from "./locales/es.json";
import ru from "./locales/ru.json";
import { DEFAULT_LOCALE, type LocaleCode, resolveLocale } from "./locales";

export type MessageTree = {
  [key: string]: string | MessageTree;
};

export type TranslateParams = Record<string, string | number>;

const catalogs: Record<LocaleCode, MessageTree> = { en, de, es, ru };

let activeLocale: LocaleCode = DEFAULT_LOCALE;

export function getCatalogLocale(): LocaleCode {
  return activeLocale;
}

export function setCatalogLocale(locale: string): LocaleCode {
  const resolved = resolveLocale(locale);
  activeLocale = resolved;
  if (typeof document !== "undefined") {
    document.documentElement.lang = resolved;
  }
  return resolved;
}

export function detectOSLocale(): string {
  if (typeof navigator !== "undefined" && navigator.language) {
    return navigator.language;
  }
  return DEFAULT_LOCALE;
}

function getNested(tree: MessageTree, key: string): string | undefined {
  const parts = key.split(".");
  let current: string | MessageTree | undefined = tree;
  for (const part of parts) {
    if (!current || typeof current !== "string") {
      current = (current as MessageTree)[part];
    } else {
      return undefined;
    }
  }
  return typeof current === "string" ? current : undefined;
}

function interpolate(template: string, params?: TranslateParams): string {
  if (!params) {
    return template;
  }
  return template.replace(/\{(\w+)\}/g, (_, name: string) => {
    const value = params[name];
    return value === undefined ? `{${name}}` : String(value);
  });
}

export function translate(key: string, params?: TranslateParams, locale = activeLocale): string {
  let value = getNested(catalogs[locale], key);
  if (value === undefined && locale !== DEFAULT_LOCALE) {
    value = getNested(catalogs[DEFAULT_LOCALE], key);
  }
  if (value === undefined) {
    return key;
  }
  return interpolate(value, params);
}

export function translatePermission(perm: string, locale = activeLocale): string {
  const lookup = (code: LocaleCode): string | undefined => {
    const extensions = catalogs[code].extensions;
    if (!extensions || typeof extensions === "string") {
      return undefined;
    }
    const labels = extensions.permission;
    if (!labels || typeof labels === "string") {
      return undefined;
    }
    const value = labels[perm];
    return typeof value === "string" ? value : undefined;
  };
  return lookup(locale) ?? lookup(DEFAULT_LOCALE) ?? perm;
}

export function flattenMessageKeys(tree: MessageTree, prefix = ""): string[] {
  const keys: string[] = [];
  for (const [name, value] of Object.entries(tree)) {
    if (name.startsWith("_")) {
      continue;
    }
    const path = prefix ? `${prefix}.${name}` : name;
    if (typeof value === "string") {
      keys.push(path);
    } else {
      keys.push(...flattenMessageKeys(value, path));
    }
  }
  return keys.sort();
}

export function getMessageByFlatKey(tree: MessageTree, flatKey: string): string | undefined {
  let found: string | undefined;
  const walk = (node: MessageTree, prefix: string) => {
    for (const [name, value] of Object.entries(node)) {
      if (name.startsWith("_")) {
        continue;
      }
      const path = prefix ? `${prefix}.${name}` : name;
      if (typeof value === "string") {
        if (path === flatKey) {
          found = value;
        }
      } else {
        walk(value, path);
      }
    }
  };
  walk(tree, "");
  return found;
}

export function listLocaleCatalogs(): Record<LocaleCode, MessageTree> {
  return catalogs;
}

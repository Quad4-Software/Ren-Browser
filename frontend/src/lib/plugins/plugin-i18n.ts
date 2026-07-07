// SPDX-License-Identifier: MIT
import type { MessageTree, TranslateParams } from "$lib/i18n/catalog";
import { DEFAULT_LOCALE, type LocaleCode, resolveLocale } from "$lib/i18n/locales";

export type PluginI18n = {
  locale: LocaleCode;
  t(key: string, params?: TranslateParams): string;
  onChange(listener: () => void): () => void;
};

type PluginCatalogSet = Partial<Record<LocaleCode, MessageTree>>;

const catalogsByPlugin = new Map<string, PluginCatalogSet>();
const instancesByPlugin = new Map<string, PluginI18n>();

function pluginLocaleUrl(pluginId: string, locale: LocaleCode): string {
  return `/_plugins/${encodeURIComponent(pluginId)}/locales/${locale}.json`;
}

async function fetchPluginCatalog(
  pluginId: string,
  locale: LocaleCode,
): Promise<MessageTree | null> {
  try {
    const resp = await fetch(pluginLocaleUrl(pluginId, locale));
    if (!resp.ok) {
      return null;
    }
    const data = (await resp.json()) as MessageTree;
    return data && typeof data === "object" ? data : null;
  } catch {
    return null;
  }
}

function getNested(tree: MessageTree, key: string): string | undefined {
  const parts = key.split(".");
  let current: string | MessageTree | undefined = tree;
  for (const part of parts) {
    if (!current || typeof current === "string") {
      return undefined;
    }
    current = (current as MessageTree)[part];
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

function translateFromCatalogs(
  catalogs: PluginCatalogSet,
  key: string,
  locale: LocaleCode,
  params?: TranslateParams,
): string {
  const primary = catalogs[locale];
  let value = primary ? getNested(primary, key) : undefined;
  if (value === undefined && locale !== DEFAULT_LOCALE) {
    const fallback = catalogs[DEFAULT_LOCALE];
    value = fallback ? getNested(fallback, key) : undefined;
  }
  if (value === undefined) {
    return key;
  }
  return interpolate(value, params);
}

export function parsePluginI18nRef(value: string): string | null {
  const trimmed = value.trim();
  if (trimmed.length < 3 || !trimmed.startsWith("%") || !trimmed.endsWith("%")) {
    return null;
  }
  return trimmed.slice(1, -1);
}

export function resolvePluginLabel(pluginId: string, title: string, locale: LocaleCode): string {
  const key = parsePluginI18nRef(title);
  if (!key) {
    return title;
  }
  const catalogs = catalogsByPlugin.get(pluginId);
  if (!catalogs) {
    return title;
  }
  const translated = translateFromCatalogs(catalogs, key, locale);
  return translated === key ? title.replaceAll("%", "") : translated;
}

export async function loadPluginCatalogs(
  pluginId: string,
  locale: LocaleCode,
): Promise<PluginCatalogSet> {
  const resolved = resolveLocale(locale);
  const localesToLoad = new Set<LocaleCode>([DEFAULT_LOCALE, resolved]);
  const catalogs: PluginCatalogSet = { ...(catalogsByPlugin.get(pluginId) ?? {}) };

  for (const code of localesToLoad) {
    const tree = await fetchPluginCatalog(pluginId, code);
    if (tree) {
      catalogs[code] = tree;
    }
  }

  catalogsByPlugin.set(pluginId, catalogs);
  return catalogs;
}

function createPluginI18nInstance(pluginId: string, locale: LocaleCode): PluginI18n {
  const listeners = new Set<() => void>();
  let activeLocale = resolveLocale(locale);

  const instance: PluginI18n = {
    get locale() {
      return activeLocale;
    },
    t(key: string, params?: TranslateParams) {
      const catalogs = catalogsByPlugin.get(pluginId) ?? {};
      return translateFromCatalogs(catalogs, key, activeLocale, params);
    },
    onChange(listener: () => void) {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
  };

  (instance as PluginI18n & { _setLocale(code: LocaleCode): void })._setLocale = (code) => {
    activeLocale = resolveLocale(code);
    for (const listener of listeners) {
      listener();
    }
  };

  instancesByPlugin.set(pluginId, instance);
  return instance;
}

export async function ensurePluginI18n(pluginId: string, locale: LocaleCode): Promise<PluginI18n> {
  const resolved = resolveLocale(locale);
  await loadPluginCatalogs(pluginId, resolved);
  const existing = instancesByPlugin.get(pluginId);
  if (existing) {
    (existing as PluginI18n & { _setLocale(code: LocaleCode): void })._setLocale(resolved);
    return existing;
  }
  return createPluginI18nInstance(pluginId, resolved);
}

export async function preloadPluginI18n(pluginIds: Iterable<string>, locale: LocaleCode) {
  const resolved = resolveLocale(locale);
  await Promise.all(
    [...new Set(pluginIds)].map((pluginId) => loadPluginCatalogs(pluginId, resolved)),
  );
}

export async function refreshPluginI18nLocales(locale: LocaleCode) {
  const resolved = resolveLocale(locale);
  const pluginIds = [...catalogsByPlugin.keys()];
  catalogsByPlugin.clear();
  await preloadPluginI18n(pluginIds, resolved);
  for (const instance of instancesByPlugin.values()) {
    (instance as PluginI18n & { _setLocale(code: LocaleCode): void })._setLocale(resolved);
  }
}

export function clearPluginI18n(pluginId?: string) {
  if (!pluginId) {
    catalogsByPlugin.clear();
    instancesByPlugin.clear();
    return;
  }
  catalogsByPlugin.delete(pluginId);
  instancesByPlugin.delete(pluginId);
}

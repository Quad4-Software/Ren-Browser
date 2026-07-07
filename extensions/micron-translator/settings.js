// SPDX-License-Identifier: MIT

const LANG_STORAGE_KEY = "renbrowser:micron-translator:langs:v1";

export const defaultSettings = {
  backend: "google",
  targetLang: "en",
  sourceLang: "auto",
  libretranslateUrl: "https://libretranslate.com",
  libretranslateApiKey: "",
};

export function cloneDefaultSettings() {
  return { ...defaultSettings, ...loadLangPrefs() };
}

export function loadLangPrefs() {
  try {
    const raw = localStorage.getItem(LANG_STORAGE_KEY);
    if (!raw) {
      return {};
    }
    const parsed = JSON.parse(raw);
    const out = {};
    if (typeof parsed.targetLang === "string" && parsed.targetLang.trim()) {
      out.targetLang = parsed.targetLang.trim();
    }
    if (typeof parsed.sourceLang === "string" && parsed.sourceLang.trim()) {
      out.sourceLang = parsed.sourceLang.trim();
    }
    return out;
  } catch {
    return {};
  }
}

export function saveLangPrefs(settings) {
  try {
    localStorage.setItem(
      LANG_STORAGE_KEY,
      JSON.stringify({
        targetLang: settings.targetLang ?? defaultSettings.targetLang,
        sourceLang: settings.sourceLang ?? defaultSettings.sourceLang,
      }),
    );
  } catch {
    // ignore storage failures
  }
}

export async function loadSettings(storage) {
  const prefs = loadLangPrefs();
  const raw = await storage.get("settings");
  if (!raw) {
    return { ...defaultSettings, ...prefs };
  }
  try {
    return { ...defaultSettings, ...prefs, ...JSON.parse(raw) };
  } catch {
    return { ...defaultSettings, ...prefs };
  }
}

export async function saveSettings(storage, settings) {
  saveLangPrefs(settings);
  await storage.set("settings", JSON.stringify(settings));
}

export function originalStorageKey(pageUrl) {
  return `original:${pageUrl}`;
}

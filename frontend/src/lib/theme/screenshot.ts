// SPDX-License-Identifier: MIT

export type ScreenshotScene = "home" | "about" | "settings" | "editor" | "discovery" | "docs";
export type ScreenshotLayout = "desktop" | "mobile";

const SCENES = new Set<ScreenshotScene>([
  "home",
  "about",
  "settings",
  "editor",
  "discovery",
  "docs",
]);
const LAYOUTS = new Set<ScreenshotLayout>(["desktop", "mobile"]);

export function screenshotLayoutFromQuery(): ScreenshotLayout | null {
  const value = new URLSearchParams(window.location.search).get("screenshot-layout");
  if (value && LAYOUTS.has(value as ScreenshotLayout)) {
    return value as ScreenshotLayout;
  }
  return null;
}

export function screenshotSceneFromQuery(): ScreenshotScene | null {
  const value = new URLSearchParams(window.location.search).get("screenshot-scene");
  if (value && SCENES.has(value as ScreenshotScene)) {
    return value as ScreenshotScene;
  }
  return null;
}

export function markScreenshotReady(): void {
  document.documentElement.dataset.screenshotReady = "true";
}

export function screenshotThemeFromQuery(): "dark" | "light" | null {
  const value = new URLSearchParams(window.location.search).get("screenshot-theme");
  if (value === "dark" || value === "light") {
    return value;
  }
  return null;
}

export function applyScreenshotQueryTheme(): "dark" | "light" | null {
  const mode = screenshotThemeFromQuery();
  if (!mode) {
    return null;
  }
  document.documentElement.dataset.theme = mode;
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) {
    meta.setAttribute("content", mode === "light" ? "#ffffff" : "#18181b");
  }
  return mode;
}

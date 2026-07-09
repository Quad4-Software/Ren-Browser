// SPDX-License-Identifier: MIT
export type ThemeMode = "dark" | "light" | "system";

export type ThemeSettings = {
  mode: ThemeMode;
  accent: string;
  fontFamily: string;
  fontSize: number;
  customTokens: Record<string, string>;
  compactToolbar: boolean;
  overlaySidebars: boolean;
};

export const defaultTheme = (): ThemeSettings => ({
  mode: "dark",
  accent: "#60a5fa",
  fontFamily: "system-ui, -apple-system, Segoe UI, sans-serif",
  fontSize: 14,
  customTokens: {},
  compactToolbar: false,
  overlaySidebars: false,
});

export function applyTheme(theme: ThemeSettings): void {
  const root = document.documentElement;
  const resolved =
    theme.mode === "system"
      ? window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light"
      : theme.mode;

  root.dataset.theme = resolved;
  root.dataset.compactToolbar = theme.compactToolbar ? "true" : "false";
  root.dataset.overlaySidebars = theme.overlaySidebars ? "true" : "false";
  root.style.setProperty("--ren-accent", theme.accent);
  root.style.setProperty("--ren-font-family", theme.fontFamily);
  root.style.setProperty("--ren-font-size", `${theme.fontSize}px`);

  for (const [key, value] of Object.entries(theme.customTokens)) {
    const token = key.startsWith("--") ? key : `--ren-${key}`;
    root.style.setProperty(token, value);
  }

  const meta = document.querySelector('meta[name="theme-color"]');
  const chromeBg = resolved === "light" ? "#ffffff" : "#18181b";
  if (meta) {
    meta.setAttribute("content", chromeBg);
  }
}

export function mobileChromeBg(theme: ThemeSettings): string {
  const resolved =
    theme.mode === "system"
      ? window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light"
      : theme.mode;
  return resolved === "light" ? "#ffffff" : "#18181b";
}

export function mobileChromeUsesLightIcons(theme: ThemeSettings): boolean {
  const resolved =
    theme.mode === "system"
      ? window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light"
      : theme.mode;
  return resolved === "dark";
}

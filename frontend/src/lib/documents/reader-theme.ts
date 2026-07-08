// SPDX-License-Identifier: MIT

export type ReaderTheme = "dark" | "light";

export type ReaderColors = {
  bg: string;
  fg: string;
  muted: string;
  border: string;
};

const READER_COLORS: Record<ReaderTheme, ReaderColors> = {
  dark: {
    bg: "#09090b",
    fg: "#f3f4f6",
    muted: "#9ca3af",
    border: "#3f3f46",
  },
  light: {
    bg: "#f8fafc",
    fg: "#111827",
    muted: "#6b7280",
    border: "#d1d5db",
  },
};

export function resolvedReaderTheme(): ReaderTheme {
  return document.documentElement.dataset.theme === "light" ? "light" : "dark";
}

export function readerColors(theme: ReaderTheme): ReaderColors {
  return READER_COLORS[theme];
}

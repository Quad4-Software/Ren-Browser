// SPDX-License-Identifier: MIT
import type { Component } from "svelte";
import {
  BookOpen,
  Compass,
  Cpu,
  Globe,
  Languages,
  Package,
  Puzzle,
  Terminal,
  Wrench,
} from "@lucide/svelte";

const pluginLucideIcons: Record<string, Component> = {
  book: BookOpen,
  "book-open": BookOpen,
  compass: Compass,
  cpu: Cpu,
  globe: Globe,
  language: Languages,
  languages: Languages,
  package: Package,
  plugin: Puzzle,
  puzzle: Puzzle,
  terminal: Terminal,
  translate: Languages,
  translator: Languages,
  wrench: Wrench,
};

export function normalizePluginIconName(name: string): string {
  return name.trim().toLowerCase().replace(/_/g, "-");
}

export function resolvePluginLucideIcon(name?: string | null): Component | null {
  if (!name?.trim()) {
    return null;
  }
  return pluginLucideIcons[normalizePluginIconName(name)] ?? null;
}

export function listPluginLucideIconNames(): string[] {
  return Object.keys(pluginLucideIcons).sort();
}

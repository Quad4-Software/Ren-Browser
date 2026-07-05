// SPDX-License-Identifier: MIT

const BLOCKED_HREF_PREFIXES = [
  "http:",
  "https:",
  "ftp:",
  "file:",
  "javascript:",
  "data:",
  "mailto:",
  "tel:",
  "blob:",
  "vbscript:",
  "ws:",
  "wss:",
] as const;

const WEB_URL_RE = /^[a-z][a-z0-9+.-]*:\/\//i;

export function isBlockedExternalHref(href: string): boolean {
  const trimmed = href.trim();
  if (!trimmed) {
    return false;
  }
  const lower = trimmed.toLowerCase();
  if (lower.startsWith("//")) {
    return true;
  }
  return BLOCKED_HREF_PREFIXES.some((prefix) => lower.startsWith(prefix));
}

export function isBlockedNavigationURL(url: string): boolean {
  const trimmed = url.trim();
  if (!trimmed) {
    return false;
  }
  if (isBlockedExternalHref(trimmed)) {
    return true;
  }
  if (!WEB_URL_RE.test(trimmed)) {
    return false;
  }
  return !trimmed.toLowerCase().startsWith("rns://");
}

export function isAllowedNavigationURL(url: string): boolean {
  const trimmed = url.trim();
  if (!trimmed) {
    return false;
  }
  return !isBlockedNavigationURL(trimmed);
}

export function blockExternalLinkPointerEvent(event: Event): boolean {
  const target = event.target;
  if (!(target instanceof Element)) {
    return false;
  }
  const anchor = target.closest("a[href]");
  if (!anchor) {
    return false;
  }
  const href = anchor.getAttribute("href");
  if (!href || !isBlockedExternalHref(href)) {
    return false;
  }
  event.preventDefault();
  event.stopPropagation();
  return true;
}

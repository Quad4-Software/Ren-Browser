// SPDX-License-Identifier: MIT

// Mesh favorites identify a node, not an exact path: a bare hash and any
// "hash:/..." URL on that node are treated as the same favorite, mirroring
// the Discovery panel's favoriteKeys logic. Non-mesh URLs (about:, docs:,
// document:, ...) only ever match themselves exactly.
const FAVORITE_NODE_HASH_RE = /^[a-f0-9]{32}$/i;

const BUILTIN_PAGE_PREFIXES = ["about:", "license:", "editor:", "config:", "settings:"] as const;

function favoriteKey(url: string): string {
  const trimmed = url.trim();
  if (!trimmed) {
    return "";
  }
  const sep = trimmed.indexOf(":/");
  const head = (sep >= 0 ? trimmed.slice(0, sep) : trimmed).trim();
  if (FAVORITE_NODE_HASH_RE.test(head)) {
    return `node:${head.toLowerCase()}`;
  }
  return `url:${trimmed}`;
}

/**
 * Returns the actually-stored favorite entry that matches `url`, preferring an
 * exact match and otherwise falling back to hash equivalence. Returns
 * undefined when `url` is not favorited.
 */
export function findFavoriteUrl(url: string, favorites: readonly string[]): string | undefined {
  const trimmed = url.trim();
  if (!trimmed) {
    return undefined;
  }
  const exact = favorites.find((fav) => fav === trimmed);
  if (exact !== undefined) {
    return exact;
  }
  const key = favoriteKey(trimmed);
  if (!key) {
    return undefined;
  }
  return favorites.find((fav) => favoriteKey(fav) === key);
}

export function isFavoriteUrl(url: string, favorites: readonly string[]): boolean {
  return findFavoriteUrl(url, favorites) !== undefined;
}

/**
 * Whether the star/favorite affordance makes sense for `url`: built-in app
 * pages (about, license, editor, config, settings, docs) are not sites and
 * cannot be favorited.
 */
export function isFavoritableUrl(url: string): boolean {
  const trimmed = url.trim();
  if (!trimmed) {
    return false;
  }
  const lower = trimmed.toLowerCase();
  if (BUILTIN_PAGE_PREFIXES.some((prefix) => lower.startsWith(prefix))) {
    return false;
  }
  return !(lower === "docs" || lower.startsWith("docs:") || lower.startsWith("docs?"));
}

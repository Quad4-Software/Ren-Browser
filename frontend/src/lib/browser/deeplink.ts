// SPDX-License-Identifier: MIT

/** Unwrap OS deeplink wrappers into an internal navigation URL. */
export function unwrapDeepLink(raw: string): string {
  const trimmed = raw.trim();
  if (!trimmed) {
    return "";
  }
  if (trimmed.includes("\0")) {
    return "";
  }
  // Reject invalid UTF-16 surrogate leftovers from bad inputs.
  try {
    encodeURIComponent(trimmed);
  } catch {
    return "";
  }

  const lower = trimmed.toLowerCase();
  if (
    lower.startsWith("http:") ||
    lower.startsWith("https:") ||
    lower.startsWith("javascript:") ||
    lower.startsWith("data:") ||
    lower.startsWith("file:") ||
    lower.startsWith("blob:") ||
    lower.startsWith("ftp:") ||
    lower.startsWith("mailto:")
  ) {
    return "";
  }

  if (lower.startsWith("renbrowser:")) {
    return unwrapRenBrowser(trimmed);
  }
  if (lower.startsWith("rns://")) {
    return trimmed;
  }
  return trimmed;
}

function unwrapRenBrowser(raw: string): string {
  let rest = raw.slice("renbrowser".length);
  if (rest.startsWith(":")) {
    rest = rest.slice(1);
  }
  rest = rest.trim();
  if (!rest) {
    return "";
  }

  if (!rest.startsWith("//")) {
    return unwrapOpaque(rest);
  }

  try {
    const parsed = new URL(raw);
    const host = parsed.hostname.toLowerCase();
    const path = parsed.pathname;
    const query = parsed.searchParams;

    if (host === "open" || (host === "" && path.replace(/\//g, "") === "open")) {
      const encoded = query.get("url") || query.get("u") || "";
      if (!encoded) {
        return "";
      }
      return unwrapDeepLink(encoded);
    }

    if (host && (path === "" || path === "/") && !parsed.search) {
      if (isBuiltin(host)) {
        return `${host}:`;
      }
      if (/^[a-f0-9]{32}$/i.test(host)) {
        return `${host.toLowerCase()}:/page/index.mu`;
      }
    }

    if (/^[a-f0-9]{32}$/i.test(host)) {
      let meshPath = path || "/page/index.mu";
      if (meshPath === "/") {
        meshPath = "/page/index.mu";
      }
      let target = `${host.toLowerCase()}:${meshPath}`;
      if (parsed.search) {
        target += parsed.search;
      }
      return target;
    }

    if (host === "rns") {
      const rnsPath = path.replace(/^\//, "");
      if (!rnsPath) {
        return "";
      }
      return unwrapDeepLink(`rns://${rnsPath}${parsed.search}`);
    }

    if (isBuiltin(host)) {
      return parsed.search ? `${host}:?${parsed.searchParams.toString()}` : `${host}:`;
    }
  } catch {
    // Fall through to opaque parsing.
  }

  return unwrapOpaque(rest.replace(/^\/\//, ""));
}

function unwrapOpaque(rest: string): string {
  const lower = rest.toLowerCase();
  if (lower.startsWith("open?") || lower.startsWith("open/?")) {
    const q = rest.slice(rest.indexOf("?") + 1);
    const params = new URLSearchParams(q);
    const encoded = params.get("url") || params.get("u") || "";
    return encoded ? unwrapDeepLink(encoded) : "";
  }
  if (lower.startsWith("url=")) {
    const params = new URLSearchParams(rest);
    const encoded = params.get("url") || "";
    return encoded ? unwrapDeepLink(encoded) : "";
  }
  if (isBuiltin(lower) || isBuiltin(lower.replace(/:$/, ""))) {
    const name = lower.replace(/:$/, "").split(/[?:]/)[0];
    if (lower.includes("?")) {
      return `${name}:?${rest.slice(rest.indexOf("?") + 1)}`;
    }
    return `${name}:`;
  }
  return unwrapDeepLink(rest);
}

function isBuiltin(name: string): boolean {
  const n = name.toLowerCase().replace(/:$/, "");
  return ["about", "license", "editor", "config", "settings", "docs", "hello"].includes(n);
}

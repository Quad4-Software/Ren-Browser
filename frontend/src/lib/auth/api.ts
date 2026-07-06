// SPDX-License-Identifier: MIT
export type AuthStatus = {
  authRequired: boolean;
  authenticated: boolean;
};

export type LoginResult = {
  ok: boolean;
  error?: string;
  blocked?: boolean;
  retryIn?: number;
};

function apiBase(): string {
  if (typeof document === "undefined") {
    return "";
  }
  const meta = document.querySelector('meta[name="ren-base-path"]');
  const base = meta?.getAttribute("content")?.trim() ?? "";
  if (!base || base === "/") {
    return "";
  }
  return base.replace(/\/$/, "");
}

export async function fetchAuthStatus(): Promise<AuthStatus> {
  try {
    const res = await fetch(`${apiBase()}/api/auth/status`, { credentials: "same-origin" });
    if (!res.ok) {
      return { authRequired: false, authenticated: true };
    }
    return (await res.json()) as AuthStatus;
  } catch {
    return { authRequired: false, authenticated: true };
  }
}

export async function login(password: string): Promise<LoginResult> {
  const res = await fetch(`${apiBase()}/api/auth/login`, {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ password }),
  });
  const data = (await res.json()) as LoginResult;
  if (!res.ok && !data.error) {
    data.error = "login failed";
  }
  return data;
}

// SPDX-License-Identifier: MIT
import { describe, expect, it, vi } from "vitest";
import { fetchAuthStatus, login } from "./api";

function jsonResponse(body: unknown, init: { ok?: boolean; status?: number } = {}): Response {
  return {
    ok: init.ok ?? true,
    status: init.status ?? 200,
    json: async () => body,
  } as Response;
}

describe("auth api", () => {
  it("returns open access when status endpoint is missing", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse(null, { ok: false })),
    );
    await expect(fetchAuthStatus()).resolves.toEqual({
      authRequired: false,
      authenticated: true,
    });
  });

  it("posts login payload", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ ok: true }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(login("secret")).resolves.toEqual({ ok: true });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/auth/login",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ password: "secret" }),
      }),
    );
  });

  it("uses base path from meta tag", async () => {
    const meta = document.createElement("meta");
    meta.name = "ren-base-path";
    meta.content = "/ren";
    document.head.appendChild(meta);

    const fetchMock = vi.fn(async () => jsonResponse({ authRequired: true, authenticated: false }));
    vi.stubGlobal("fetch", fetchMock);

    await fetchAuthStatus();
    expect(fetchMock).toHaveBeenCalledWith(
      "/ren/api/auth/status",
      expect.objectContaining({ credentials: "same-origin" }),
    );

    document.head.removeChild(meta);
  });

  it("surfaces login errors from server", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({ ok: false, error: "invalid password" }, { ok: false, status: 401 }),
      ),
    );
    await expect(login("bad")).resolves.toEqual({
      ok: false,
      error: "invalid password",
    });
  });

  it("reports brute-force blocks", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse(
          {
            ok: false,
            blocked: true,
            retryIn: 120,
            error: "too many failed attempts",
          },
          { ok: false, status: 429 },
        ),
      ),
    );
    await expect(login("bad")).resolves.toMatchObject({
      ok: false,
      blocked: true,
      retryIn: 120,
    });
  });
});

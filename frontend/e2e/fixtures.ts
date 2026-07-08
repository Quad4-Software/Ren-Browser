// SPDX-License-Identifier: MIT
import { test as base, expect, type Page } from "@playwright/test";
import { e2eBaseURL, startServer, type ServerHandle } from "./helpers/server";

type Fixtures = {
  server: ServerHandle;
};

export const test = base.extend<Fixtures>({
  server: async ({}, use) => {
    const server = await startServer();
    await use(server);
    await server.stop();
  },
});

export { expect };

export async function openScene(
  page: Page,
  scene: string,
  opts?: { layout?: string; theme?: string; wait?: string },
): Promise<void> {
  const layout = opts?.layout ?? "desktop";
  const theme = opts?.theme ?? "dark";
  const wait = opts?.wait ?? "header.chrome";
  const params = new URLSearchParams({
    "screenshot-theme": theme,
    "screenshot-scene": scene,
    "screenshot-layout": layout,
  });
  await page.goto(`${e2eBaseURL}?${params}`, { waitUntil: "networkidle" });
  await page.evaluate((mode) => {
    document.documentElement.dataset.theme = mode;
  }, theme);
  await page
    .waitForSelector('[data-screenshot-ready="true"]', { timeout: 60_000 })
    .catch(() => page.waitForSelector(wait, { timeout: 60_000 }));
  await page.waitForTimeout(400);
}

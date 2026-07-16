// SPDX-License-Identifier: MIT
import { defineConfig, devices } from "@playwright/test";

const port = Number(process.env.REN_BROWSER_E2E_PORT || 19284);
const host = process.env.REN_BROWSER_E2E_HOST || "127.0.0.1";
const baseURL = `http://${host}:${port}/`;

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  timeout: 90_000,
  expect: {
    timeout: 20_000,
    toHaveScreenshot: {
      maxDiffPixelRatio: 0.02,
      animations: "disabled",
    },
  },
  reporter: process.env.CI ? [["github"], ["list"]] : "list",
  use: {
    baseURL,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
        // Match frontend/scripts/capture-screenshots.mjs desktop baselines (not Desktop Chrome's 1280x720).
        viewport: { width: 1280, height: 800 },
      },
    },
  ],
});

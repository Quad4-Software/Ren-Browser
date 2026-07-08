// SPDX-License-Identifier: MIT
import AxeBuilder from "@axe-core/playwright";
import { expect, openScene, test } from "./fixtures";

test.describe("critical paths", () => {
  test("loads home chrome and navigates to about", async ({ page, server }) => {
    void server;
    await openScene(page, "home");

    await expect(page.locator("header.chrome")).toBeVisible();
    await expect(page.locator(".url-input")).toBeVisible();
    await expect(page.locator(".new-tab")).toBeVisible();

    await page.locator(".url-input").fill("about:");
    await page.locator(".url-input").press("Enter");
    await expect(page.locator("article.about-page")).toBeVisible({ timeout: 30_000 });
  });

  test("opens a new tab from the tab bar", async ({ page, server }) => {
    void server;
    await openScene(page, "home");

    const tabsBefore = await page.locator(".tabbar .tab").count();
    await page.locator(".new-tab").click();
    await expect(page.locator(".tabbar .tab")).toHaveCount(tabsBefore + 1);
  });

  test("opens settings side pane", async ({ page, server }) => {
    void server;
    await openScene(page, "settings", { wait: ".side-pane .settings" });
    await expect(page.locator(".side-pane .settings")).toBeVisible();
  });

  test("opens downloads menu", async ({ page, server }) => {
    void server;
    await openScene(page, "home");
    await page.getByRole("button", { name: "Downloads" }).click();
    await expect(page.locator(".menu .page-btn, .menu .folder-btn").first()).toBeVisible({
      timeout: 10_000,
    });
  });
});

test.describe("accessibility", () => {
  test("home scene has no serious axe violations", async ({ page, server }) => {
    void server;
    await openScene(page, "home");
    const results = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa"])
      // Tab close controls are nested interactive by design today.
      .disableRules(["nested-interactive"])
      .analyze();
    const serious = results.violations.filter(
      (v) => v.impact === "critical" || v.impact === "serious",
    );
    expect(serious, JSON.stringify(serious, null, 2)).toEqual([]);
  });
});

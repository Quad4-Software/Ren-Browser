// SPDX-License-Identifier: MIT
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { PNG } from "pngjs";
import pixelmatch from "pixelmatch";
import { expect, openScene, test } from "./fixtures";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");

const scenes = [
  { name: "home", wait: "header.chrome" },
  { name: "about", wait: "article.about-page" },
  { name: "settings", wait: ".side-pane .settings" },
  { name: "docs", wait: "article.docs-page" },
] as const;

function comparePNG(actual: Buffer, expectedPath: string, maxDiffRatio = 0.03): number {
  const expected = PNG.sync.read(readFileSync(expectedPath));
  const actualPng = PNG.sync.read(actual);
  if (expected.width !== actualPng.width || expected.height !== actualPng.height) {
    throw new Error(
      `size mismatch for ${expectedPath}: expected ${expected.width}x${expected.height}, got ${actualPng.width}x${actualPng.height}`,
    );
  }
  const diff = new PNG({ width: expected.width, height: expected.height });
  const mismatched = pixelmatch(
    expected.data,
    actualPng.data,
    diff.data,
    expected.width,
    expected.height,
    { threshold: 0.1 },
  );
  const ratio = mismatched / (expected.width * expected.height);
  if (ratio > maxDiffRatio) {
    throw new Error(
      `${expectedPath}: ${(ratio * 100).toFixed(2)}% pixels differ (limit ${(maxDiffRatio * 100).toFixed(1)}%)`,
    );
  }
  return ratio;
}

test.describe("visual regression", () => {
  for (const scene of scenes) {
    test(`desktop dark ${scene.name}`, async ({ page, server }) => {
      void server;
      await openScene(page, scene.name, { wait: scene.wait, theme: "dark", layout: "desktop" });
      const shot = await page.screenshot({ fullPage: false, type: "png" });
      const baseline = path.join(root, "screenshots", "desktop", "dark", `${scene.name}.png`);
      // Allow modest font/AA drift across hosts; CI step is advisory until baselines are CI-captured.
      expect(() => comparePNG(shot, baseline, 0.08)).not.toThrow();
    });
  }
});

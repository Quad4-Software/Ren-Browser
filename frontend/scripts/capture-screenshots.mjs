// SPDX-License-Identifier: MIT
import { spawn } from "node:child_process";
import { mkdir, mkdtemp, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { chromium } from "playwright";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const binName = process.platform === "win32" ? "renbrowser-server.exe" : "renbrowser-server";
const serverBin = path.join(root, "bin", binName);
const outRoot = path.join(root, "screenshots");
const port = Number(process.env.REN_BROWSER_SCREENSHOT_PORT || 19283);
const host = "127.0.0.1";
const baseURL = `http://${host}:${port}/`;
const themes = ["dark", "light"];

const layouts = [
  {
    id: "desktop",
    viewport: { width: 1280, height: 800 },
    scenes: [
      { name: "home", scene: "home", wait: "header.chrome" },
      { name: "about", scene: "about", wait: "article.about-page" },
      { name: "settings", scene: "settings", wait: ".side-pane .settings" },
      { name: "editor", scene: "editor", wait: "#micron-source" },
      { name: "discovery", scene: "discovery", wait: ".side-pane h2" },
    ],
  },
  {
    id: "mobile",
    viewport: { width: 390, height: 844 },
    scenes: [
      { name: "home", scene: "home", wait: ".mobile-ui .mobile-nav" },
      { name: "about", scene: "about", wait: "article.about-page" },
      { name: "settings", scene: "settings", wait: ".mobile-panel .settings" },
      { name: "editor", scene: "editor", wait: "#micron-source" },
      { name: "discovery", scene: "discovery", wait: ".mobile-panel h2" },
    ],
  },
];

async function waitForServer(url, timeoutMs = 60_000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(url, { redirect: "follow" });
      if (res.ok) {
        return;
      }
    } catch {
      // retry until the server is ready
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`server did not become ready at ${url}`);
}

function startServer(profileHome) {
  return spawn(serverBin, ["--host", host, "--port", String(port)], {
    cwd: root,
    env: {
      ...process.env,
      HOME: profileHome,
      REN_BROWSER_PUBLIC_MODE: "1",
    },
    stdio: ["ignore", "pipe", "pipe"],
  });
}

async function applyTheme(page, mode) {
  await page.evaluate((theme) => {
    document.documentElement.dataset.theme = theme;
    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) {
      meta.setAttribute("content", theme === "light" ? "#ffffff" : "#18181b");
    }
  }, mode);
}

async function waitForScene(page, waitSelector) {
  await page
    .waitForSelector('[data-screenshot-ready="true"]', { timeout: 60_000 })
    .catch(() => page.waitForSelector(waitSelector, { timeout: 60_000 }));
  await page.waitForTimeout(500);
}

async function captureShot(browser, layout, scene, mode) {
  const dir = path.join(outRoot, layout.id, mode);
  await mkdir(dir, { recursive: true });

  const page = await browser.newPage({ viewport: layout.viewport });
  const params = new URLSearchParams({
    "screenshot-theme": mode,
    "screenshot-scene": scene.scene,
    "screenshot-layout": layout.id,
  });
  await page.goto(`${baseURL}?${params}`, { waitUntil: "networkidle" });
  await applyTheme(page, mode);
  await waitForScene(page, scene.wait);

  const file = path.join(dir, `${scene.name}.png`);
  await page.screenshot({ path: file, fullPage: false });
  await page.close();
  return file;
}

async function main() {
  const profileHome = await mkdtemp(path.join(os.tmpdir(), "renbrowser-screenshot-"));
  const server = startServer(profileHome);
  let serverLog = "";

  server.stdout?.on("data", (chunk) => {
    serverLog += chunk.toString();
  });
  server.stderr?.on("data", (chunk) => {
    serverLog += chunk.toString();
  });

  const browser = await chromium.launch({ headless: true });

  try {
    await waitForServer(baseURL);
    const files = [];
    for (const layout of layouts) {
      for (const scene of layout.scenes) {
        for (const mode of themes) {
          files.push(await captureShot(browser, layout, scene, mode));
        }
      }
    }
    for (const file of files) {
      process.stdout.write(`${file}\n`);
    }
  } catch (err) {
    if (serverLog.trim()) {
      process.stderr.write(`${serverLog.trim()}\n`);
    }
    throw err;
  } finally {
    await browser.close();
    server.kill("SIGTERM");
    await new Promise((resolve) => {
      if (server.exitCode !== null) {
        resolve();
        return;
      }
      server.once("exit", () => resolve());
      setTimeout(() => {
        try {
          server.kill("SIGKILL");
        } catch {
          // already exited
        }
        resolve();
      }, 5_000);
    });
    await rm(profileHome, { recursive: true, force: true });
  }
}

main().catch((err) => {
  process.stderr.write(`${err instanceof Error ? err.message : String(err)}\n`);
  process.exit(1);
});

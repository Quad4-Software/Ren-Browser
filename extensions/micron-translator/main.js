// SPDX-License-Identifier: MIT
import {
  cloneDefaultSettings,
  loadSettings,
  originalStorageKey,
  saveLangPrefs,
  saveSettings,
} from "./settings.js";

const PLUGIN_ID = "renbrowser.micron-translator";
const PLUGIN_VERSION = "1.0.0";
const PLUGIN_BACKEND = "wasm";

let handlersRegistered = false;
let panelCtx = null;
let panelSettings = cloneDefaultSettings();
let panelBusy = false;

function backendNeedsNetwork(settings) {
  const backend = settings?.backend || "google";
  return backend === "google" || backend === "libretranslate";
}

async function wasmTranslate(ctx, body, settings) {
  if (backendNeedsNetwork(settings) && !ctx.capabilities?.networkFetch) {
    throw new Error(ctx.i18n.t("toast.networkRequired"));
  }
  if (!ctx.wasm) {
    throw new Error("plugin wasm backend is unavailable");
  }
  const resp = await ctx.wasm.call("translate_micron", { body, settings });
  if (resp?.error) {
    throw new Error(resp.error);
  }
  return resp?.body ?? "";
}

async function rememberOriginal(ctx, page) {
  if (!page?.url || !page.raw) {
    return;
  }
  const key = originalStorageKey(page.url);
  const existing = await ctx.storage.get(key);
  if (!existing) {
    await ctx.storage.set(key, page.raw);
  }
}

async function translateActivePage(ctx) {
  const page = ctx.content.getActivePage();
  if (page.contentType !== "micron" || !page.raw) {
    ctx.ui.showToast(ctx.i18n.t("toast.openMicronPage"));
    return;
  }
  await rememberOriginal(ctx, page);
  const settings = await loadSettings(ctx.storage);
  ctx.ui.showToast(ctx.i18n.t("toast.translating"));
  const translated = await wasmTranslate(ctx, page.raw, settings);
  const path = page.path || "/page.mu";
  const rendered = await ctx.content.renderRaw(path, translated);
  ctx.content.updateActivePage({
    html: rendered.html,
    raw: rendered.raw,
    contentType: rendered.contentType,
    path,
  });
  ctx.ui.showToast(ctx.i18n.t("toast.translated"));
}

async function restoreActivePage(ctx) {
  const page = ctx.content.getActivePage();
  if (!page.url) {
    return;
  }
  const key = originalStorageKey(page.url);
  const original = await ctx.storage.get(key);
  if (!original) {
    ctx.ui.showToast(ctx.i18n.t("toast.noOriginal"));
    return;
  }
  const path = page.path || "/page.mu";
  const rendered = await ctx.content.renderRaw(path, original);
  ctx.content.updateActivePage({
    html: rendered.html,
    raw: rendered.raw,
    contentType: rendered.contentType,
    path,
  });
  ctx.ui.showToast(ctx.i18n.t("toast.restored"));
}

function renderPanel(root, t) {
  root.innerHTML = `
    <article class="translator-panel">
      <style>
        .translator-panel {
          display: grid;
          gap: 0.65rem;
          font-size: 0.9rem;
          flex: 1;
          min-height: 0;
        }
        .translator-panel .title { margin: 0; font-size: 1rem; }
        .translator-panel .hint { margin: 0; color: var(--ren-muted); line-height: 1.4; }
        .translator-panel label { display: grid; gap: 0.25rem; }
        .translator-panel label[data-libre-field] { display: none; }
        .translator-panel.libre label[data-libre-field] { display: grid; }
        .translator-panel input:not(.ren-input),
        .translator-panel button {
          font: inherit;
          border-radius: calc(var(--ren-radius) + 2px);
          border: 1px solid var(--ren-border);
          background: var(--ren-input-bg);
          color: var(--ren-fg);
          padding: 0.35rem 0.5rem;
        }
        .translator-panel .ren-input,
        .translator-panel .ren-select {
          padding: 0.35rem 0.5rem;
          font-size: inherit;
        }
        .translator-panel .ren-select { width: 100%; }
        .translator-panel .actions { display: grid; gap: 0.4rem; }
        .translator-panel .status { margin: 0; min-height: 1.2rem; color: var(--ren-muted); }
        .translator-panel .meta {
          margin: auto 0 0;
          padding-top: 0.35rem;
          color: var(--ren-muted);
          font-size: 0.75rem;
          letter-spacing: 0.02em;
        }
      </style>
      <h2 class="title">${t("panel.title")}</h2>
      <p class="hint">${t("panel.hint")}</p>
      <label>
        <span>${t("panel.backend")}</span>
        <select data-field="backend" class="ren-select">
          <option value="google">${t("panel.backendGoogle")}</option>
          <option value="libretranslate">${t("panel.backendLibre")}</option>
        </select>
      </label>
      <label>
        <span>${t("panel.targetLang")}</span>
        <input data-field="targetLang" class="ren-input" type="text" maxlength="16" autocomplete="off" />
      </label>
      <label>
        <span>${t("panel.sourceLang")}</span>
        <input data-field="sourceLang" class="ren-input" type="text" maxlength="16" autocomplete="off" placeholder="${t("panel.sourceLangPlaceholder")}" />
      </label>
      <label data-libre-field>
        <span>${t("panel.libreUrl")}</span>
        <input data-field="libretranslateUrl" class="ren-input" type="url" autocomplete="off" />
      </label>
      <label data-libre-field>
        <span>${t("panel.libreApiKey")}</span>
        <input data-field="libretranslateApiKey" class="ren-input" type="password" autocomplete="off" />
      </label>
      <div class="actions">
        <button type="button" data-action="translate">${t("panel.translate")}</button>
        <button type="button" data-action="restore">${t("panel.restore")}</button>
        <button type="button" data-action="save">${t("panel.save")}</button>
      </div>
      <p class="status" data-status></p>
      <p class="meta">v${PLUGIN_VERSION} · ${PLUGIN_BACKEND}</p>
    </article>
  `;
  return root.querySelector(".translator-panel");
}

function readPanelSettings(panel) {
  const read = (name) => panel.querySelector(`[data-field="${name}"]`)?.value?.trim() ?? "";
  return {
    backend: read("backend") || "google",
    targetLang: read("targetLang") || "en",
    sourceLang: read("sourceLang") || "auto",
    libretranslateUrl: read("libretranslateUrl") || "https://libretranslate.com",
    libretranslateApiKey: read("libretranslateApiKey"),
  };
}

function applyPanelSettings(panel, settings) {
  for (const [key, value] of Object.entries(settings)) {
    const el = panel.querySelector(`[data-field="${key}"]`);
    if (el) {
      el.value = value ?? "";
    }
  }
  panel.classList.toggle("libre", settings.backend === "libretranslate");
}

function setPanelStatus(panel, message) {
  const el = panel.querySelector("[data-status]");
  if (el) {
    el.textContent = message;
  }
}

async function bindPanel(panel, ctx) {
  const t = (key) => ctx.i18n.t(key);
  panelSettings = await loadSettings(ctx.storage);
  applyPanelSettings(panel, panelSettings);

  panel.querySelector('[data-field="backend"]')?.addEventListener("change", (event) => {
    const backend = event.target.value;
    panel.classList.toggle("libre", backend === "libretranslate");
  });

  for (const field of ["targetLang", "sourceLang"]) {
    panel.querySelector(`[data-field="${field}"]`)?.addEventListener("change", () => {
      saveLangPrefs(readPanelSettings(panel));
    });
  }

  panel.querySelector('[data-action="save"]')?.addEventListener("click", async () => {
    panelSettings = readPanelSettings(panel);
    await saveSettings(ctx.storage, panelSettings);
    setPanelStatus(panel, t("panel.settingsSaved"));
  });

  panel.querySelector('[data-action="translate"]')?.addEventListener("click", async () => {
    if (panelBusy) {
      return;
    }
    panelBusy = true;
    setPanelStatus(panel, t("panel.translating"));
    try {
      panelSettings = readPanelSettings(panel);
      await saveSettings(ctx.storage, panelSettings);
      await translateActivePage(ctx);
      setPanelStatus(panel, t("panel.translationApplied"));
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setPanelStatus(panel, message);
      ctx.ui.showToast(message);
    } finally {
      panelBusy = false;
    }
  });

  panel.querySelector('[data-action="restore"]')?.addEventListener("click", async () => {
    if (panelBusy) {
      return;
    }
    panelBusy = true;
    try {
      await restoreActivePage(ctx);
      setPanelStatus(panel, t("panel.originalRestored"));
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      setPanelStatus(panel, message);
      ctx.ui.showToast(message);
    } finally {
      panelBusy = false;
    }
  });
}

export function activate(ctx) {
  panelCtx = ctx;
  if (handlersRegistered) {
    return;
  }
  handlersRegistered = true;

  ctx.subscriptions.add(
    ctx.events.on("page:loaded", async (data) => {
      const page = data;
      if (page?.contentType === "micron" && page?.raw && page?.url) {
        await rememberOriginal(ctx, {
          url: page.url,
          raw: page.raw,
        });
      }
    }),
  );

  ctx.subscriptions.add(
    ctx.events.on(`plugin:${PLUGIN_ID}:command:translate-page`, () => {
      void translateActivePage(ctx).catch((err) => {
        ctx.ui.showToast(err instanceof Error ? err.message : String(err));
      });
    }),
  );

  ctx.subscriptions.add(
    ctx.events.on(`plugin:${PLUGIN_ID}:command:restore-original`, () => {
      void restoreActivePage(ctx).catch((err) => {
        ctx.ui.showToast(err instanceof Error ? err.message : String(err));
      });
    }),
  );
}

export function deactivate() {
  panelCtx = null;
  handlersRegistered = false;
}

export async function mount(el) {
  if (!panelCtx) {
    return;
  }
  const panel = renderPanel(el, (key) => panelCtx.i18n.t(key));
  await bindPanel(panel, panelCtx);
}

// SPDX-License-Identifier: MIT
import { Events } from "@wailsio/runtime";
import { RenderRaw } from "../../../bindings/renbrowser/internal/app/browserservice.js";
import {
  AddTrustedPublisher,
  DisablePlugin,
  EnablePlugin,
  GetContributions,
  GetPluginStorage,
  InstallPluginFromDir,
  InstallPluginFromWasm,
  InstallPluginFromZip,
  InvokeCommand,
  ListPlugins,
  PluginFetch,
  PluginWasmCall,
  PreviewPluginInstallFromDir,
  PreviewPluginInstallFromWasm,
  PreviewPluginInstallFromZip,
  SetPluginStorage,
  UninstallPlugin,
} from "../../../bindings/renbrowser/internal/app/pluginhost.js";
import type { PluginInstallPreview as BindingPluginInstallPreview } from "../../../bindings/renbrowser/internal/app/models.js";
import type {
  ActivePageSnapshot,
  ContributionsSnapshot,
  PluginContext,
  PluginHTTPRequest,
  PluginModule,
  RenderedPageSnapshot,
} from "./api-types.js";
import type { PluginI18n } from "./plugin-i18n.js";
import { formatBindingError, toBindingError } from "$lib/browser/binding-errors.js";

export async function listPlugins() {
  return ListPlugins();
}

export async function enablePlugin(id: string) {
  return EnablePlugin(id);
}

export async function disablePlugin(id: string) {
  return DisablePlugin(id);
}

export type PluginSignatureInfo = {
  present: boolean;
  valid: boolean;
  signer?: string;
  signerName?: string;
  trusted: boolean;
  error?: string;
};

export type PluginSecurityFinding = {
  id: string;
  severity: string;
  message: string;
};

export type PluginSecurityAssessment = {
  riskLevel: string;
  score: number;
  findings: PluginSecurityFinding[];
};

export type PluginInstallPreview = {
  id: string;
  name: string;
  version: string;
  description?: string;
  permissions: string[];
  networkEndpoints: string[];
  requiresNetworkFetch: boolean;
  signature: PluginSignatureInfo;
  security: PluginSecurityAssessment;
  i18nLocales: string[];
};

function emptySecurityAssessment(): PluginSecurityAssessment {
  return { riskLevel: "low", score: 0, findings: [] };
}

function emptyPluginSignature(): PluginSignatureInfo {
  return { present: false, valid: false, trusted: false };
}

function normalizeSecurityAssessment(
  raw: BindingPluginInstallPreview["security"] | undefined,
): PluginSecurityAssessment {
  if (!raw) {
    return emptySecurityAssessment();
  }
  return {
    riskLevel: raw.riskLevel ?? "low",
    score: raw.score ?? 0,
    findings: raw.findings ?? [],
  };
}

function normalizeInstallPreview(raw: BindingPluginInstallPreview): PluginInstallPreview {
  return {
    id: raw.id ?? "",
    name: raw.name ?? "",
    version: raw.version ?? "",
    description: raw.description,
    permissions: raw.permissions ?? [],
    networkEndpoints: raw.networkEndpoints ?? [],
    requiresNetworkFetch: raw.requiresNetworkFetch ?? false,
    signature: raw.signature ?? emptyPluginSignature(),
    security: normalizeSecurityAssessment(raw.security),
    i18nLocales: raw.i18nLocales ?? [],
  };
}

export async function previewInstallFromZip(path: string): Promise<PluginInstallPreview> {
  return normalizeInstallPreview(await PreviewPluginInstallFromZip(path));
}

export async function previewInstallFromDir(path: string): Promise<PluginInstallPreview> {
  return normalizeInstallPreview(await PreviewPluginInstallFromDir(path));
}

export async function previewInstallFromWasm(path: string): Promise<PluginInstallPreview> {
  return normalizeInstallPreview(await PreviewPluginInstallFromWasm(path));
}

export type PluginInstallChoices = {
  dontShowAgain: boolean;
  trustPublisher: boolean;
  grantedPermissions: string[];
};

export async function installFromZip(path: string, granted: string[] = []) {
  return InstallPluginFromZip(path, granted);
}

export async function installFromDir(path: string, granted: string[] = []) {
  return InstallPluginFromDir(path, granted);
}

export async function installFromWasm(path: string, granted: string[] = []) {
  return InstallPluginFromWasm(path, granted);
}

export async function trustPublisher(identity: string, name = "") {
  return AddTrustedPublisher(identity, name);
}

export async function uninstallPlugin(id: string) {
  return UninstallPlugin(id);
}

export async function getContributions(): Promise<ContributionsSnapshot> {
  const raw = await GetContributions();
  return {
    panels: (raw.panels ?? []).map((p) => ({
      pluginId: p.pluginId,
      id: p.id,
      title: p.title,
      icon: p.icon,
      entry: p.entry,
      location: p.location,
    })),
    commands: (raw.commands ?? []).map((c) => ({
      pluginId: c.pluginId,
      commandId: c.id,
      title: c.title,
      keybind: c.keybind,
    })),
    devtools: (raw.devtools ?? []).map((d) => ({
      pluginId: d.pluginId,
      id: d.id,
      title: d.title,
      entry: d.entry,
    })),
    urlSchemes: (raw.urlSchemes ?? []).map((s) => ({
      pluginId: s.pluginId,
      scheme: s.scheme,
      handler: s.handler,
    })),
  };
}

export async function invokeCommand(
  pluginId: string,
  commandId: string,
  args: Record<string, string> = {},
) {
  return InvokeCommand(pluginId, commandId, args);
}

export function createPluginContext(
  pluginId: string,
  opts: {
    getCurrentURL: () => string;
    navigate: (url: string) => void;
    showToast: (message: string) => void;
    getActivePage: () => ActivePageSnapshot;
    updateActivePage: (patch: Partial<ActivePageSnapshot>) => void;
    networkFetch?: boolean;
    wasmBackend?: boolean;
    i18n: PluginI18n;
  },
): PluginContext {
  const disposables: Array<{ dispose(): void }> = [];
  const ctx: PluginContext = {
    pluginId,
    subscriptions: {
      add(disposable) {
        disposables.push(disposable);
      },
    },
    storage: {
      async get(key) {
        const value = await GetPluginStorage(pluginId, key);
        return value || null;
      },
      async set(key, value) {
        await SetPluginStorage(pluginId, key, value);
      },
    },
    navigation: {
      getCurrentURL: opts.getCurrentURL,
      navigate: opts.navigate,
    },
    content: {
      getActivePage: opts.getActivePage,
      updateActivePage: opts.updateActivePage,
      async renderRaw(path, raw): Promise<RenderedPageSnapshot> {
        try {
          const page = await RenderRaw(path, raw);
          return {
            html: page.html ?? "",
            contentType: page.contentType ?? "",
            raw: page.raw ?? raw,
            pageFg: page.pageFg,
            pageBg: page.pageBg,
          };
        } catch (err) {
          throw toBindingError(err, "Could not render page");
        }
      },
    },
    events: {
      on(event, fn) {
        const full = event.includes(":") ? event : `plugin:${pluginId}:${event}`;
        const cancel = Events.On(full, (payload) => {
          void (async () => {
            try {
              fn((payload as { data?: unknown }).data ?? payload);
            } catch (err) {
              const { reportPluginFailure } = await import("./plugin-errors.js");
              await reportPluginFailure(pluginId, `event:${event}`, err);
            }
          })();
        });
        return { dispose: () => cancel() };
      },
      emit(event, data) {
        const full = event.includes(":") ? event : `plugin:${pluginId}:${event}`;
        Events.Emit(full, data);
      },
    },
    ui: {
      showToast: opts.showToast,
      formatError: (err: unknown, fallback = "Extension failed") =>
        formatBindingError(err, fallback),
    },
    capabilities: {
      networkFetch: opts.networkFetch === true,
      wasmBackend: opts.wasmBackend === true,
    },
    i18n: opts.i18n,
  };
  if (opts.networkFetch) {
    ctx.network = {
      async fetch(req: PluginHTTPRequest) {
        try {
          const resp = await PluginFetch(pluginId, {
            method: req.method ?? "",
            url: req.url,
            headers: req.headers ?? {},
            body: req.body ?? "",
          });
          return {
            statusCode: resp.statusCode ?? 0,
            body: resp.body ?? "",
          };
        } catch (err) {
          throw toBindingError(err, "Network request failed");
        }
      },
    };
  }
  if (opts.wasmBackend) {
    ctx.wasm = {
      async call(exportName, input) {
        let raw: string;
        try {
          raw = await PluginWasmCall(pluginId, exportName, JSON.stringify(input ?? {}));
        } catch (err) {
          throw toBindingError(err, "Extension call failed");
        }
        if (!raw) {
          return {};
        }
        let parsed: Record<string, unknown>;
        try {
          parsed = JSON.parse(raw) as Record<string, unknown>;
        } catch {
          return { body: raw };
        }
        if (typeof parsed.error === "string" && parsed.error.trim()) {
          throw new Error(formatBindingError(parsed.error, "Extension call failed"));
        }
        return parsed;
      },
    };
  }
  return ctx;
}

export type { PluginModule };

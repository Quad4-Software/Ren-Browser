// SPDX-License-Identifier: MIT
import { translate } from "$lib/i18n/catalog";

/**
 * Micron inline images arrive from the parser as deferred placeholders:
 *
 *   <div class="mu-image" data-mu-image-path=":/media/x.png" ...>
 *     <span class="mu-image-meta">alt + optional size hint</span>
 *     <span class="mu-image-actions"><a data-mu-image-action="load">...</a></span>
 *     <img class="mu-image-output" hidden>
 *   </div>
 *
 * Nothing is fetched until the user opts in (per image, per node, or via the
 * global "always" mode). This module hydrates those placeholders: it resolves
 * the node hash, enforces policy, fetches bytes through the injected loader,
 * and swaps in a blob-backed img. All network bytes stay on the mesh; the
 * backend additionally enforces a byte cap and a raster-only content sniff.
 */

export type MicronImagesMode = "off" | "ask" | "always";
export type MicronImageNodePolicy = "always" | "never";

export type MicronImageFetchResult = {
  data: string;
  mime: string;
  bytes: number;
};

export type MicronImageLabels = {
  load: string;
  allowNode: string;
  disabled: string;
  blocked: string;
  tooLarge: string;
  loading: string;
  failed: string;
};

export type MicronImageOptions = {
  pageNodeHash: string;
  mode: MicronImagesMode;
  nodePolicies: Readonly<Record<string, string>>;
  fetchImage: (url: string) => Promise<MicronImageFetchResult>;
  onNodePolicy?: (nodeHash: string, policy: MicronImageNodePolicy | null) => void;
  labels?: Partial<MicronImageLabels>;
};

export type MicronImageHandle = {
  teardown: () => void;
};

export const MICRON_IMAGE_MAX_BYTES = 32 * 1024 * 1024;
export const MICRON_IMAGE_AUTOLOAD_MAX = 50;
const MICRON_IMAGE_FETCH_CONCURRENCY = 3;

const nodeHashRe = /^[a-f0-9]{32}$/i;

export function micronImageLabels(): MicronImageLabels {
  return {
    load: translate("content.imageLoad"),
    allowNode: translate("content.imageAllowNode"),
    disabled: translate("content.imageDisabled"),
    blocked: translate("content.imageBlocked"),
    tooLarge: translate("content.imageTooLarge"),
    loading: translate("content.imageLoading"),
    failed: translate("content.imageFailed"),
  };
}

export function normalizeMicronImagesMode(value: unknown): MicronImagesMode {
  return value === "off" || value === "always" ? value : "ask";
}

/**
 * Effective policy for a node. A per-node policy beats the global mode;
 * "never" also beats a global "always".
 */
export function micronImageNodeDecision(
  mode: MicronImagesMode,
  nodePolicies: Readonly<Record<string, string>>,
  nodeHash: string | null,
): "always" | "ask" | "never" {
  if (nodeHash) {
    const policy = nodePolicies[nodeHash.toLowerCase()];
    if (policy === "always") {
      return "always";
    }
    if (policy === "never") {
      return "never";
    }
  }
  if (mode === "off") {
    return "never";
  }
  return mode;
}

/**
 * Resolves data-mu-image-path ("<hash>:/media/x.png" or ":/media/x.png")
 * against the current page node. Returns null for anything that is not a
 * well-formed node image reference.
 */
export function resolveMicronImageURL(
  path: string,
  pageNodeHash: string,
): { url: string; nodeHash: string } | null {
  const raw = (path ?? "").trim();
  let nodeHash: string;
  let rest: string;
  const m = raw.match(/^([a-f0-9]{32}):(\/.*)$/i);
  if (m) {
    nodeHash = m[1].toLowerCase();
    rest = m[2];
  } else if (raw.startsWith(":/")) {
    nodeHash = pageNodeHash.toLowerCase();
    rest = raw.slice(1);
  } else {
    return null;
  }
  if (!nodeHashRe.test(nodeHash)) {
    return null;
  }
  if (!rest.startsWith("/media/") && !rest.startsWith("/file/")) {
    return null;
  }
  if (rest.includes("..")) {
    return null;
  }
  return { url: `${nodeHash}:${rest}`, nodeHash };
}

function base64ToBlob(data: string, mime: string): Blob {
  const bin = atob(data);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) {
    bytes[i] = bin.charCodeAt(i);
  }
  return new Blob([bytes], { type: mime });
}

export function attachMicronImages(
  root: HTMLElement,
  options: MicronImageOptions,
): MicronImageHandle {
  const labels = { ...micronImageLabels(), ...options.labels };
  const objectURLs = new Map<string, HTMLImageElement>();
  const loading = new Set<HTMLElement>();
  let disposed = false;
  let autoLoaded = 0;
  let inFlight = 0;
  const queue: HTMLElement[] = [];

  function setState(holder: HTMLElement, state: string) {
    holder.setAttribute("data-mu-image-state", state);
  }

  function setActionText(holder: HTMLElement, text: string) {
    const action = holder.querySelector<HTMLElement>("[data-mu-image-action='load']");
    if (action) {
      action.textContent = text;
    }
  }

  function pumpQueue() {
    while (inFlight < MICRON_IMAGE_FETCH_CONCURRENCY && queue.length > 0) {
      const next = queue.shift();
      if (next) {
        void loadImage(next);
      }
    }
  }

  function enqueue(holder: HTMLElement) {
    if (disposed || loading.has(holder)) {
      return;
    }
    queue.push(holder);
    pumpQueue();
  }

  async function loadImage(holder: HTMLElement) {
    if (
      disposed ||
      loading.has(holder) ||
      holder.getAttribute("data-mu-image-state") === "loaded"
    ) {
      return;
    }
    const resolved = resolveMicronImageURL(
      holder.getAttribute("data-mu-image-path") ?? "",
      options.pageNodeHash,
    );
    if (!resolved) {
      setState(holder, "error");
      setActionText(holder, labels.failed);
      return;
    }
    if (
      micronImageNodeDecision(options.mode, options.nodePolicies, resolved.nodeHash) === "never"
    ) {
      if (options.mode === "off") {
        setState(holder, "disabled");
        setActionText(holder, labels.disabled);
      } else {
        setState(holder, "blocked");
        setActionText(holder, labels.blocked);
      }
      return;
    }
    const sizeHint = Number(holder.getAttribute("data-mu-image-s") ?? "0");
    if (sizeHint > MICRON_IMAGE_MAX_BYTES) {
      setState(holder, "too-large");
      setActionText(holder, labels.tooLarge);
      return;
    }
    loading.add(holder);
    inFlight++;
    setState(holder, "loading");
    setActionText(holder, labels.loading);
    try {
      const result = await options.fetchImage(resolved.url);
      if (!result || !result.data || result.bytes > MICRON_IMAGE_MAX_BYTES) {
        throw new Error("invalid image response");
      }
      const img = holder.querySelector<HTMLImageElement>(".mu-image-output");
      if (!img) {
        throw new Error("missing image output element");
      }
      const url = URL.createObjectURL(base64ToBlob(result.data, result.mime || "image/png"));
      objectURLs.set(url, img);
      img.src = url;
      img.alt = holder.getAttribute("data-mu-image-alt") ?? "";
      img.decoding = "async";
      img.hidden = false;
      setState(holder, "loaded");
    } catch {
      if (!disposed) {
        setState(holder, "error");
        setActionText(holder, labels.failed);
      }
    } finally {
      loading.delete(holder);
      inFlight--;
      pumpQueue();
    }
  }

  function allowNode(holder: HTMLElement) {
    const resolved = resolveMicronImageURL(
      holder.getAttribute("data-mu-image-path") ?? "",
      options.pageNodeHash,
    );
    if (!resolved || !options.onNodePolicy) {
      return;
    }
    options.onNodePolicy(resolved.nodeHash, "always");
    for (const el of root.querySelectorAll<HTMLElement>(".mu-image")) {
      if (el.getAttribute("data-mu-image-state") === "loaded") {
        continue;
      }
      const other = resolveMicronImageURL(
        el.getAttribute("data-mu-image-path") ?? "",
        options.pageNodeHash,
      );
      if (other && other.nodeHash === resolved.nodeHash) {
        enqueue(el);
      }
    }
  }

  function activateAction(el: HTMLElement) {
    if (el.getAttribute("aria-disabled") === "true") {
      return;
    }
    const action = el.getAttribute("data-mu-image-action");
    const holder = el.closest<HTMLElement>(".mu-image");
    if (!holder || loading.has(holder)) {
      return;
    }
    if (holder.getAttribute("data-mu-image-state") === "loaded") {
      return;
    }
    if (action === "load") {
      void loadImage(holder);
    } else if (action === "node-allow") {
      allowNode(holder);
    }
  }

  const onClick = (event: Event) => {
    const target = event.target;
    if (!(target instanceof HTMLElement)) {
      return;
    }
    const action = target.closest<HTMLElement>("[data-mu-image-action]");
    if (!action || !root.contains(action)) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    activateAction(action);
  };

  const onKeydown = (event: KeyboardEvent) => {
    if (event.key !== "Enter" && event.key !== " ") {
      return;
    }
    const target = event.target;
    if (!(target instanceof HTMLElement)) {
      return;
    }
    const action = target.closest<HTMLElement>("[data-mu-image-action]");
    if (!action || !root.contains(action)) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    activateAction(action);
  };

  root.addEventListener("click", onClick);
  root.addEventListener("keydown", onKeydown);

  for (const holder of root.querySelectorAll<HTMLElement>(".mu-image")) {
    if (holder.getAttribute("data-mu-image-state") === "loaded") {
      continue;
    }
    const resolved = resolveMicronImageURL(
      holder.getAttribute("data-mu-image-path") ?? "",
      options.pageNodeHash,
    );
    const decision = micronImageNodeDecision(
      options.mode,
      options.nodePolicies,
      resolved ? resolved.nodeHash : null,
    );
    const actions = holder.querySelector<HTMLElement>(".mu-image-actions");
    const sizeHint = Number(holder.getAttribute("data-mu-image-s") ?? "0");

    if (decision === "never") {
      setState(holder, options.mode === "off" ? "disabled" : "blocked");
      setActionText(holder, options.mode === "off" ? labels.disabled : labels.blocked);
      const action = holder.querySelector<HTMLElement>("[data-mu-image-action='load']");
      if (action) {
        action.setAttribute("aria-disabled", "true");
        action.setAttribute("tabindex", "-1");
      }
      continue;
    }
    if (sizeHint > MICRON_IMAGE_MAX_BYTES) {
      setState(holder, "too-large");
      setActionText(holder, labels.tooLarge);
      continue;
    }

    if (options.mode === "ask" && resolved && options.onNodePolicy) {
      const policy = options.nodePolicies[resolved.nodeHash];
      if (
        policy !== "always" &&
        actions &&
        !actions.querySelector("[data-mu-image-action='node-allow']")
      ) {
        const allow = document.createElement("a");
        allow.className = "mu-image-action mu-image-action-node";
        allow.setAttribute("data-mu-image-action", "node-allow");
        allow.setAttribute("role", "button");
        allow.setAttribute("tabindex", "0");
        allow.textContent = labels.allowNode;
        actions.appendChild(allow);
      }
    }

    if (decision === "always" && autoLoaded < MICRON_IMAGE_AUTOLOAD_MAX) {
      autoLoaded++;
      enqueue(holder);
    }
  }

  return {
    teardown: () => {
      disposed = true;
      queue.length = 0;
      root.removeEventListener("click", onClick);
      root.removeEventListener("keydown", onKeydown);
      // Only revoke blob URLs whose img is gone. Re-attach teardown (policy or
      // mode change) keeps the DOM alive, so revoking those URLs would break
      // already-loaded images.
      for (const [url, img] of objectURLs) {
        if (!img.isConnected) {
          URL.revokeObjectURL(url);
        }
      }
      objectURLs.clear();
    },
  };
}

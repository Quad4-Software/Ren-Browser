// SPDX-License-Identifier: MIT
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  attachMicronImages,
  micronImageNodeDecision,
  resolveMicronImageURL,
  MICRON_IMAGE_MAX_BYTES,
} from "./images";
import { renderClientMicronPage } from "./render-page";

const NODE = "abb3ebcd03cb2388a838e70c001291f9";
const OTHER = "11111111111111111111111111111111";

const PNG_B64 = "iVBORw0KGgoAAAANSUhEUg==";

function holder(path = `:/media/x.png`, extra = ""): string {
  return (
    `<div class="mu-image" data-mu-image-url="${path}" data-mu-image-path="${path}" ` +
    `data-mu-image-alt="alt text" role="img" aria-label="alt text" ${extra}>` +
    `<span class="mu-image-meta"><span class="mu-image-alt">alt text</span></span>` +
    `<span class="mu-image-actions"><a class="mu-image-action" data-mu-image-action="load" role="button" tabindex="0">Load image</a></span>` +
    `<img class="mu-image-output" alt="alt text" hidden=""></div>`
  );
}

function makeRoot(html: string): HTMLElement {
  const root = document.createElement("div");
  // eslint-disable-next-line no-unsanitized/property -- test fixtures are built from literals
  root.innerHTML = html;
  document.body.appendChild(root);
  return root;
}

function fetchOK() {
  return vi.fn(async () => ({ data: PNG_B64, mime: "image/png", bytes: 12 }));
}

describe("JS fallback image detection", () => {
  const url = `${NODE}:/page/index.mu`;

  it("emits a deferred mu-image placeholder for media links with fields", () => {
    const html = renderClientMicronPage(url, "`[River valley`:/media/harbour.png`w=400]", "js");
    expect(html).toContain('class="mu-image"');
    expect(html).toContain('data-mu-image-path=":/media/harbour.png"');
    expect(html).toContain('data-mu-image-w="400"');
    expect(html).toContain('data-mu-image-action="load"');
    expect(html).toContain('class="mu-image-output"');
  });

  it("emits a placeholder for /file/ webp links marked img=1", () => {
    const html = renderClientMicronPage(url, "`[sticker`:/file/x.webp`img=1]", "js");
    expect(html).toContain('data-mu-image-path=":/file/x.webp"');
  });

  it("keeps field-less media links and non-media links as plain anchors", () => {
    const html = renderClientMicronPage(url, "`[a`:/media/x.png]", "js");
    expect(html).not.toContain("mu-image");
    expect(html).toContain("Mu-nl");

    const pageLink = renderClientMicronPage(url, "`[a`:/page/x.mu`img=1]", "js");
    expect(pageLink).not.toContain("mu-image");
  });

  it("escapes hostile alt text and rejects traversal paths", () => {
    const html = renderClientMicronPage(url, "`[<img onerror=x>`:/media/../etc/x.png`img=1]", "js");
    expect(html).not.toContain("mu-image");
    expect(html).not.toContain("<img onerror");

    const unsafe = renderClientMicronPage(url, "`[a`:/media/x.png`img=1;k=<script>]", "js");
    expect(unsafe).not.toContain("<script");
  });
});

describe("resolveMicronImageURL", () => {
  it("resolves same-node and cross-node paths", () => {
    expect(resolveMicronImageURL(":/media/x.png", NODE)).toEqual({
      url: `${NODE}:/media/x.png`,
      nodeHash: NODE,
    });
    expect(resolveMicronImageURL(`${OTHER}:/media/x.png`, NODE)).toEqual({
      url: `${OTHER}:/media/x.png`,
      nodeHash: OTHER,
    });
    expect(resolveMicronImageURL(`${OTHER}:/file/x.webp`, NODE)?.nodeHash).toBe(OTHER);
  });

  it("rejects non-image and malformed paths", () => {
    expect(resolveMicronImageURL(":/page/x.mu", NODE)).toBeNull();
    expect(resolveMicronImageURL("http://evil/x.png", NODE)).toBeNull();
    expect(resolveMicronImageURL(":/media/../x.png", NODE)).toBeNull();
    expect(resolveMicronImageURL(":/media/x.png", "notahash")).toBeNull();
    expect(resolveMicronImageURL("", NODE)).toBeNull();
  });
});

describe("micronImageNodeDecision", () => {
  it("per-node policy overrides the global mode", () => {
    expect(micronImageNodeDecision("ask", { [NODE]: "always" }, NODE)).toBe("always");
    expect(micronImageNodeDecision("always", { [NODE]: "never" }, NODE)).toBe("never");
    expect(micronImageNodeDecision("off", { [NODE]: "always" }, NODE)).toBe("always");
  });

  it("falls back to the global mode", () => {
    expect(micronImageNodeDecision("ask", {}, NODE)).toBe("ask");
    expect(micronImageNodeDecision("off", {}, NODE)).toBe("never");
    expect(micronImageNodeDecision("always", {}, NODE)).toBe("always");
    expect(micronImageNodeDecision("always", {}, null)).toBe("always");
  });
});

describe("attachMicronImages", () => {
  let root: HTMLElement | null = null;

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    root?.remove();
    root = null;
  });

  it("does not fetch until the user clicks Load image", async () => {
    const fetchImage = fetchOK();
    root = makeRoot(holder());
    attachMicronImages(root, {
      pageNodeHash: NODE,
      mode: "ask",
      nodePolicies: {},
      fetchImage,
    });
    expect(fetchImage).not.toHaveBeenCalled();
    const img = root.querySelector<HTMLImageElement>(".mu-image-output");
    expect(img?.hidden).toBe(true);
    expect(img?.src ?? "").toBe("");
  });

  it("loads the image on click and swaps in a blob URL", async () => {
    const fetchImage = fetchOK();
    root = makeRoot(holder());
    attachMicronImages(root, {
      pageNodeHash: NODE,
      mode: "ask",
      nodePolicies: {},
      fetchImage,
    });
    const action = root.querySelector<HTMLElement>("[data-mu-image-action='load']");
    action?.click();
    await vi.waitFor(() => {
      expect(root?.querySelector(".mu-image")?.getAttribute("data-mu-image-state")).toBe("loaded");
    });
    expect(fetchImage).toHaveBeenCalledWith(`${NODE}:/media/x.png`);
    const img = root.querySelector<HTMLImageElement>(".mu-image-output");
    expect(img?.hidden).toBe(false);
    expect(img?.src.startsWith("blob:")).toBe(true);
  });

  it("activates with Enter and Space for keyboard users", async () => {
    const fetchImage = fetchOK();
    root = makeRoot(holder());
    attachMicronImages(root, {
      pageNodeHash: NODE,
      mode: "ask",
      nodePolicies: {},
      fetchImage,
    });
    const action = root.querySelector<HTMLElement>("[data-mu-image-action='load']")!;
    action.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
    await vi.waitFor(() => {
      expect(fetchImage).toHaveBeenCalledTimes(1);
    });
  });

  it("auto-loads when the node is allowed or mode is always", async () => {
    const fetchImage = fetchOK();
    root = makeRoot(holder());
    attachMicronImages(root, {
      pageNodeHash: NODE,
      mode: "ask",
      nodePolicies: { [NODE]: "always" },
      fetchImage,
    });
    await vi.waitFor(() => {
      expect(fetchImage).toHaveBeenCalledTimes(1);
    });
  });

  it("never fetches when the mode is off", async () => {
    const fetchImage = fetchOK();
    root = makeRoot(holder());
    attachMicronImages(root, {
      pageNodeHash: NODE,
      mode: "off",
      nodePolicies: {},
      fetchImage,
    });
    const action = root.querySelector<HTMLElement>("[data-mu-image-action='load']");
    action?.click();
    await new Promise((r) => setTimeout(r, 20));
    expect(fetchImage).not.toHaveBeenCalled();
    expect(root.querySelector(".mu-image")?.getAttribute("data-mu-image-state")).toBe("disabled");
  });

  it("blocks images for a never policy node even when mode is always", async () => {
    const fetchImage = fetchOK();
    root = makeRoot(holder(`${NODE}:/media/x.png`));
    attachMicronImages(root, {
      pageNodeHash: NODE,
      mode: "always",
      nodePolicies: { [NODE]: "never" },
      fetchImage,
    });
    const action = root.querySelector<HTMLElement>("[data-mu-image-action='load']");
    action?.click();
    await new Promise((r) => setTimeout(r, 20));
    expect(fetchImage).not.toHaveBeenCalled();
    expect(root.querySelector(".mu-image")?.getAttribute("data-mu-image-state")).toBe("blocked");
  });

  it("refuses images that advertise more than the byte cap", async () => {
    const fetchImage = fetchOK();
    root = makeRoot(holder(":/media/big.png", `data-mu-image-s="${MICRON_IMAGE_MAX_BYTES + 1}"`));
    attachMicronImages(root, {
      pageNodeHash: NODE,
      mode: "always",
      nodePolicies: {},
      fetchImage,
    });
    await new Promise((r) => setTimeout(r, 20));
    expect(fetchImage).not.toHaveBeenCalled();
    expect(root.querySelector(".mu-image")?.getAttribute("data-mu-image-state")).toBe("too-large");
  });

  it("marks the placeholder when the fetch fails", async () => {
    const fetchImage = vi.fn(async () => {
      throw new Error("no path to node");
    });
    root = makeRoot(holder());
    attachMicronImages(root, {
      pageNodeHash: NODE,
      mode: "ask",
      nodePolicies: {},
      fetchImage,
    });
    root.querySelector<HTMLElement>("[data-mu-image-action='load']")?.click();
    await vi.waitFor(() => {
      expect(root?.querySelector(".mu-image")?.getAttribute("data-mu-image-state")).toBe("error");
    });
  });

  it("offers an always-allow action that loads every image on the node", async () => {
    const fetchImage = fetchOK();
    const onNodePolicy = vi.fn();
    root = makeRoot(holder() + holder(":/media/y.png"));
    attachMicronImages(root, {
      pageNodeHash: NODE,
      mode: "ask",
      nodePolicies: {},
      fetchImage,
      onNodePolicy,
    });
    const allow = root.querySelector<HTMLElement>("[data-mu-image-action='node-allow']");
    expect(allow).not.toBeNull();
    allow?.click();
    expect(onNodePolicy).toHaveBeenCalledWith(NODE, "always");
    await vi.waitFor(() => {
      expect(fetchImage).toHaveBeenCalledTimes(2);
    });
  });
});

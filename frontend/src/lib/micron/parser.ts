// SPDX-License-Identifier: MIT
// @ts-nocheck
import DOMPurify from "dompurify";
import BaseMicronParser from "micron-parser";

const ALLOWED_URI_REGEXP =
  /^(?:(?:(?:f|ht)tps?|mailto|tel|callto|cid|xmpp|nomadnetwork|lxmf):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i;

// The upstream micron-parser.js base class references a global DOMPurify in its
// constructor and in inherited helpers, so expose the bundled instance there.
if (typeof globalThis.DOMPurify === "undefined") {
  globalThis.DOMPurify = DOMPurify;
}

function escapeHtmlForFallback(text) {
  if (text == null) return "";
  return String(text)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

/**
 * Extends micron-parser-js with Nomad Network rendering needs:
 * partial includes, overlay style stripping, wide CJK monospace cells, and lxmf/nomadnetwork in DOMPurify.
 */
export default class MicronParser extends BaseMicronParser {
  constructor(darkTheme = true, enableForceMonospace = true) {
    super(darkTheme, enableForceMonospace);
    if (this.enableForceMonospace) {
      const existing = document.getElementById("micron-monospace-styles");
      if (existing) {
        existing.remove();
      }
      this.injectMonospaceStyles();
    }
  }

  static get PARTIAL_LINE_REGEX() {
    // eslint-disable-next-line security/detect-unsafe-regex -- fixed pattern, bounded input (single line)
    return /^`\{([a-f0-9]{32}):([^`}]*)(?:`(\d+)(?:`([^}]*))?)?\}$/;
  }

  // Image detection mirrors micron-parser-go v1.2.0 image.go: only links with
  // fields can be images, only /media/ (raster ext) and /file/*.webp paths
  // qualify, and alt text is required.
  static IMAGE_MAX_ALT_LEN = 240;
  static IMAGE_MEDIA_RE = /\.(webp|png|jpe?g|bmp|gif|tiff)$/i;
  static IMAGE_HASH_RE = /^[a-f0-9]{32}$/i;
  static IMAGE_KEY_RE = /^[a-zA-Z0-9_.-]*$/;
  static IMAGE_PROFILE_RE = /^[a-zA-Z0-9_-]*$/;

  static clampImageNumber(v, limit) {
    const n = Number(String(v).trim());
    if (!Number.isFinite(n) || n <= 0) return 0;
    return Math.min(Math.floor(n), limit);
  }

  static sanitizeImageString(v, maxLen, re) {
    let s = String(v ?? "").trim();
    if ([...s].length > maxLen) s = [...s].slice(0, maxLen).join("");
    if (re && !re.test(s)) return "";
    return s;
  }

  static parseImageOptions(fields) {
    const opts = { img: false, w: 0, h: 0, s: 0, k: "", a: "left", profile: "" };
    for (const raw of fields) {
      if (!raw) continue;
      for (const part of String(raw).split(";")) {
        const idx = part.indexOf("=");
        if (idx <= 0) continue;
        const k = part.slice(0, idx).trim().toLowerCase();
        const v = part.slice(idx + 1).trim();
        switch (k) {
          case "img":
            opts.img = v === "1" || v.toLowerCase() === "true" || v.toLowerCase() === "yes";
            break;
          case "w":
            opts.w = MicronParser.clampImageNumber(v, 8192);
            break;
          case "h":
            opts.h = MicronParser.clampImageNumber(v, 8192);
            break;
          case "s":
            opts.s = MicronParser.clampImageNumber(v, 100 * 1024 * 1024);
            break;
          case "k":
            opts.k = MicronParser.sanitizeImageString(v, 64, MicronParser.IMAGE_KEY_RE);
            break;
          case "a": {
            const lower = v.toLowerCase();
            opts.a =
              lower === "center" || lower === "c"
                ? "center"
                : lower === "right" || lower === "r"
                  ? "right"
                  : "left";
            break;
          }
          case "profile":
            opts.profile = MicronParser.sanitizeImageString(v, 32, MicronParser.IMAGE_PROFILE_RE);
            break;
        }
      }
    }
    return opts;
  }

  static extractImageFilePath(rawURL) {
    if (!rawURL) return "";
    let url = String(rawURL);
    if (url.toLowerCase().startsWith("nomadnetwork://")) {
      url = url.slice("nomadnetwork://".length);
    }
    for (const cut of ["`", "?", "#"]) {
      const i = url.indexOf(cut);
      if (i >= 0) url = url.slice(0, i);
    }
    url = url.trim();

    let path = url;
    let hash = "";
    const sep = url.indexOf(":/");
    if (sep >= 0) {
      hash = url.slice(0, sep);
      path = url.slice(sep + 2);
      if (hash !== "" && !MicronParser.IMAGE_HASH_RE.test(hash)) return "";
    } else if (url.startsWith(":")) {
      path = url.slice(1);
    }

    if (path.startsWith("media/")) {
      if (!MicronParser.IMAGE_MEDIA_RE.test(path)) return "";
    } else if (path.startsWith("file/")) {
      if (!path.toLowerCase().endsWith(".webp")) return "";
    } else {
      return "";
    }

    if (path.includes("..")) return "";
    for (const ch of path) {
      if (ch < " " || '<>"|?*'.includes(ch)) return "";
    }
    return hash !== "" ? `${hash}:/${path}` : `:/${path}`;
  }

  /**
   * Upgrades micron-parser-js link elements into deferred .mu-image
   * placeholders, matching micron-parser-go writeImage output so both render
   * paths feed the same hydration code.
   */
  static upgradeImageLink(anchor) {
    const destination = anchor.getAttribute("data-destination") || "";
    const fieldsAttr = anchor.getAttribute("data-fields") || "";
    const backtick = destination.indexOf("`");
    const optionFields = [];
    if (backtick >= 0) {
      for (const f of destination.slice(backtick + 1).split("|")) {
        optionFields.push(f);
      }
    }
    if (fieldsAttr) {
      for (const f of fieldsAttr.split("|")) {
        optionFields.push(f);
      }
    }
    if (optionFields.length === 0) {
      return false;
    }
    const rawPath = (backtick >= 0 ? destination.slice(0, backtick) : destination).trim();
    const opts = MicronParser.parseImageOptions(optionFields);
    const imagePath = MicronParser.extractImageFilePath(rawPath);
    const withoutHash = imagePath.includes(":/") ? imagePath.split(":/")[1] : imagePath;
    const isMedia = withoutHash.startsWith("media/") || withoutHash.startsWith("/media/");
    if (!opts.img && !isMedia) {
      return false;
    }
    const alt = (anchor.textContent || "").trim().slice(0, MicronParser.IMAGE_MAX_ALT_LEN);
    if (!imagePath || !alt) {
      return false;
    }

    const doc = anchor.ownerDocument;
    const div = doc.createElement("div");
    div.className = "mu-image";
    div.setAttribute("data-mu-image-url", rawPath);
    div.setAttribute("data-mu-image-path", imagePath);
    div.setAttribute("data-mu-image-alt", alt);
    if (opts.w > 0) div.setAttribute("data-mu-image-w", String(opts.w));
    if (opts.h > 0) div.setAttribute("data-mu-image-h", String(opts.h));
    if (opts.s > 0) div.setAttribute("data-mu-image-s", String(opts.s));
    if (opts.k) div.setAttribute("data-mu-image-k", opts.k);
    if (opts.a) div.setAttribute("data-mu-image-a", opts.a);
    if (opts.profile) div.setAttribute("data-mu-image-profile", opts.profile);
    div.setAttribute("role", "img");
    div.setAttribute("aria-label", alt);
    const styleParts = [];
    if (opts.w > 0) styleParts.push(`width:${opts.w}px`);
    if (opts.h > 0) styleParts.push(`min-height:${opts.h}px`);
    if (styleParts.length) div.setAttribute("style", styleParts.join(";"));

    const meta = doc.createElement("span");
    meta.className = "mu-image-meta";
    const altSpan = doc.createElement("span");
    altSpan.className = "mu-image-alt";
    altSpan.textContent = alt;
    meta.appendChild(altSpan);
    if (opts.s > 0) {
      const sizeSpan = doc.createElement("span");
      sizeSpan.className = "mu-image-size";
      sizeSpan.textContent = MicronParser.formatImageSize(opts.s);
      meta.appendChild(doc.createTextNode(" "));
      meta.appendChild(sizeSpan);
    }

    const actions = doc.createElement("span");
    actions.className = "mu-image-actions";
    const action = doc.createElement("a");
    action.className = "mu-image-action";
    action.setAttribute("data-mu-image-action", "load");
    action.setAttribute("role", "button");
    action.setAttribute("tabindex", "0");
    action.textContent = "Load image";
    actions.appendChild(action);

    const img = doc.createElement("img");
    img.className = "mu-image-output";
    img.setAttribute("alt", alt);
    img.setAttribute("hidden", "");

    div.appendChild(meta);
    div.appendChild(actions);
    div.appendChild(img);
    anchor.replaceWith(div);
    return true;
  }

  static upgradeImageLinks(root) {
    if (!root || !root.querySelectorAll) {
      return;
    }
    const anchors = [];
    if (root.matches && root.matches("a.Mu-nl[data-destination]")) {
      anchors.push(root);
    }
    for (const a of root.querySelectorAll("a.Mu-nl[data-destination]")) {
      anchors.push(a);
    }
    for (const a of anchors) {
      try {
        MicronParser.upgradeImageLink(a);
      } catch {
        /* leave as link */
      }
    }
  }

  static formatImageSize(bytes) {
    if (bytes < 0) return "";
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${Math.round((bytes / 1024) * 10) / 10} kB`;
    return `${Math.round((bytes / (1024 * 1024)) * 10) / 10} MB`;
  }

  static isWideMonospaceCell(segment) {
    if (!segment) return false;
    if (
      /\p{Script=Han}|\p{Script=Hiragana}|\p{Script=Katakana}|\p{Script=Hangul}|\p{Script=Bopomofo}/u.test(
        segment,
      )
    ) {
      return true;
    }
    const cp = segment.codePointAt(0);
    if (cp >= 0x3000 && cp <= 0x303f) return true;
    if (cp >= 0xff01 && cp <= 0xff5e) return true;
    if (cp >= 0xffe0 && cp <= 0xffe6) return true;
    return false;
  }

  /**
   * When false, forceMonospace can render a whole word in one span (Latin/Cyrillic/etc.),
   * avoiding one DOM node per character (critical for large pages and resize performance).
   */
  static lineNeedsPerCharCells(line) {
    if (!line) {
      return false;
    }
    for (const char of line) {
      if (MicronParser.isWideMonospaceCell(char)) {
        return true;
      }
      const cp = char.codePointAt(0);
      if (cp >= 0x2500 && cp <= 0x257f) {
        return true;
      }
      if (cp >= 0x2580 && cp <= 0x259f) {
        return true;
      }
    }
    return false;
  }

  static stripOverlayStyles(html) {
    if (typeof html !== "string") return html;
    const dangerousProps = ["zindex", "inset", "top", "left", "right", "bottom", "transform"];
    return html.replace(/(\s)style="([^"]*)"/g, (match, space, styleValue) => {
      const declarations = styleValue.split(";").filter(Boolean);
      const safe = declarations.filter((decl) => {
        const colon = decl.indexOf(":");
        if (colon <= 0) return false;
        const rawProp = decl.slice(0, colon).trim();
        const prop = rawProp.toLowerCase().replace(/-/g, "");
        const val = decl
          .slice(colon + 1)
          .trim()
          .toLowerCase();
        if (prop === "position" && (val === "fixed" || val === "sticky")) return false;
        if (dangerousProps.includes(prop)) return false;
        if (prop === "width" && /100v[wh]/.test(val)) return false;
        if (prop === "height" && /100v[hw]/.test(val)) return false;
        return true;
      });
      const out = safe.join("; ").trim();
      return out ? `${space}style="${out}"` : "";
    });
  }

  static sanitizeRenderedMicronHtml(html) {
    if (html == null) {
      return "";
    }
    const s = typeof html === "string" ? html : String(html);
    try {
      const sanitized = DOMPurify.sanitize(s, {
        USE_PROFILES: { html: true },
        ALLOWED_URI_REGEXP,
      });
      try {
        return MicronParser.stripOverlayStyles(sanitized);
      } catch (e) {
        console.warn("MicronParser: stripOverlayStyles failed", e);
        return sanitized;
      }
    } catch (error) {
      console.warn(
        "DOMPurify is not installed or sanitization failed. Include dompurify or check the build.",
        error,
      );
      return `<p style="color: red;">DOMPurify is not installed or sanitization failed.</p>`;
    }
  }

  /**
   * Split Micron source into blocks for WASM conversion, keeping MeshChat partial-include lines on the JS path.
   */
  static splitMicronMarkupWasmSegments(markup) {
    if (markup == null) {
      return [];
    }
    const lines = String(markup).split("\n");
    const segments = [];
    let buf = [];
    for (const line of lines) {
      const trimmed = line.trim();
      if (MicronParser.PARTIAL_LINE_REGEX.test(trimmed)) {
        if (buf.length) {
          segments.push({ type: "mu", text: buf.join("\n") });
          buf = [];
        }
        segments.push({ type: "partial", line });
      } else {
        buf.push(line);
      }
    }
    if (buf.length) {
      segments.push({ type: "mu", text: buf.join("\n") });
    }
    return segments;
  }

  injectMonospaceStyles() {
    if (typeof document === "undefined") {
      return;
    }
    if (document.getElementById("micron-monospace-styles")) {
      return;
    }

    const styleEl = document.createElement("style");
    styleEl.id = "micron-monospace-styles";

    styleEl.textContent = `
            .Mu-nl {
                cursor: pointer;
            }
            .Mu-mnt {
                display: inline-block;
                box-sizing: border-box;
                min-width: 1ch;
                width: 1ch;
                max-width: 1ch;
                text-align: center;
                white-space: pre;
                text-decoration: inherit;
                vertical-align: baseline;
                line-height: 1.25;
            }
            .Mu-mnt-full {
                display: inline-block;
                box-sizing: border-box;
                min-width: 2ch;
                width: 2ch;
                max-width: 2ch;
                text-align: center;
                white-space: pre;
                text-decoration: inherit;
                vertical-align: baseline;
                line-height: 1.25;
            }
            .Mu-mws {
                text-decoration: inherit;
                display: inline-flex;
                flex-wrap: wrap;
                align-items: baseline;
                column-gap: 0;
                row-gap: 0;
                gap: 0;
            }
            .Mu-mnt-group {
                display: inline;
                font-family: inherit;
                white-space: pre-wrap;
                overflow-wrap: anywhere;
                word-break: break-word;
                text-decoration: inherit;
                vertical-align: baseline;
                line-height: 1.25;
            }
        `;
    document.head.appendChild(styleEl);
  }

  convertMicronToHtmlWasmHybrid(markup, partialContents = {}) {
    const mc = globalThis.micronConvert;
    const segments = MicronParser.splitMicronMarkupWasmSegments(markup);
    let html = "";
    for (const seg of segments) {
      if (seg.type === "mu") {
        html += MicronParser.sanitizeRenderedMicronHtml(
          mc(seg.text, this.darkTheme, this.enableForceMonospace),
        );
      } else {
        html += this._convertMicronToHtmlJs(seg.line + "\n", partialContents);
      }
    }
    return html;
  }

  _convertMicronToHtmlJs(markup, partialContents = {}) {
    const build = () => {
      let html = "";

      let headerColors = { fg: null, bg: null };
      try {
        headerColors = this.parseHeaderTags(markup);
      } catch (e) {
        console.warn("MicronParser: parseHeaderTags failed", e);
      }

      const plainStyle = this.SELECTED_STYLES?.plain || {
        fg: this.DEFAULT_FG_DARK,
        bg: this.DEFAULT_BG,
      };
      const defaultFg = headerColors.fg || plainStyle.fg;
      const defaultBg = headerColors.bg || plainStyle.bg;

      let state = {
        literal: false,
        depth: 0,
        fg_color: defaultFg,
        bg_color: defaultBg,
        formatting: {
          bold: false,
          underline: false,
          italic: false,
          strikethrough: false,
        },
        default_align: "left",
        align: "left",
        default_fg: defaultFg,
        default_bg: defaultBg,
        radio_groups: {},
        partialIndex: 0,
      };

      const lines = markup.split("\n");

      for (let line of lines) {
        let lineOutput;
        try {
          lineOutput = this.parseLine(line, state);
        } catch (e) {
          console.warn("MicronParser: parseLine failed", e);
          html += `<span class="mu-line-parse-fallback" style="white-space:pre-wrap">${escapeHtmlForFallback(line)}</span><br>`;
          continue;
        }
        if (lineOutput && lineOutput.length > 0) {
          for (let el of lineOutput) {
            try {
              if (el.classList && el.classList.contains("mu-partial")) {
                const id = el.getAttribute("data-partial-id");
                if (id && partialContents[id]) {
                  html += partialContents[id];
                } else {
                  html += el.outerHTML;
                }
              } else {
                MicronParser.upgradeImageLinks(el);
                html += el.outerHTML;
              }
            } catch (e) {
              console.warn("MicronParser: line output serialization failed", e);
              html += `<span class="mu-line-parse-fallback" style="white-space:pre-wrap">${escapeHtmlForFallback(line)}</span><br>`;
              break;
            }
          }
        } else if (lineOutput && lineOutput.length === 0) {
          // skip
        } else {
          html += "<br>";
        }
      }

      return MicronParser.sanitizeRenderedMicronHtml(html);
    };

    try {
      return build();
    } catch (e) {
      console.warn("MicronParser: convertMicronToHtml failed", e);
      const escaped = escapeHtmlForFallback(markup);
      try {
        return DOMPurify.sanitize(
          `<pre class="mu-parse-fallback" style="white-space:pre-wrap">${escaped}</pre>`,
          {
            USE_PROFILES: { html: true },
            ALLOWED_URI_REGEXP,
          },
        );
      } catch {
        return `<pre class="mu-parse-fallback" style="white-space:pre-wrap">${escaped}</pre>`;
      }
    }
  }

  convertMicronToHtml(markup, partialContents = {}, options = {}) {
    if (markup == null) return "";
    if (typeof markup !== "string") markup = String(markup);

    const wantWasm = options.useWasm === true && typeof globalThis.micronConvert === "function";
    if (wantWasm) {
      try {
        return this.convertMicronToHtmlWasmHybrid(markup, partialContents);
      } catch (e) {
        console.warn("MicronParser: WASM Micron conversion failed, using JS parser", e);
      }
    }

    return this._convertMicronToHtmlJs(markup, partialContents);
  }

  convertMicronToFragment(markup) {
    if (markup == null) {
      return document.createDocumentFragment();
    }
    if (typeof markup !== "string") markup = String(markup);

    try {
      const fragment = document.createDocumentFragment();

      let headerColors = { fg: null, bg: null };
      try {
        headerColors = this.parseHeaderTags(markup);
      } catch (e) {
        console.warn("MicronParser: parseHeaderTags failed", e);
      }

      const plainStyle = this.SELECTED_STYLES?.plain || {
        fg: this.DEFAULT_FG_DARK,
        bg: this.DEFAULT_BG,
      };
      const defaultFg = headerColors.fg || plainStyle.fg;
      const defaultBg = headerColors.bg || plainStyle.bg;

      let state = {
        literal: false,
        depth: 0,
        fg_color: defaultFg,
        bg_color: defaultBg,
        formatting: {
          bold: false,
          underline: false,
          italic: false,
          strikethrough: false,
        },
        default_align: "left",
        align: "left",
        default_fg: defaultFg,
        default_bg: defaultBg,
        radio_groups: {},
        partialIndex: 0,
      };

      const lines = markup.split("\n");

      for (let line of lines) {
        let sanitizedLine = line;
        try {
          sanitizedLine = DOMPurify.sanitize(line, {
            USE_PROFILES: { html: true },
            ALLOWED_URI_REGEXP,
          });
        } catch (e) {
          console.warn("MicronParser: line sanitize failed", e);
        }
        let lineOutput;
        try {
          lineOutput = this.parseLine(sanitizedLine, state);
        } catch (e) {
          console.warn("MicronParser: parseLine failed", e);
          const fallback = document.createElement("span");
          fallback.className = "mu-line-parse-fallback";
          fallback.style.whiteSpace = "pre-wrap";
          fallback.textContent = line;
          fragment.appendChild(fallback);
          fragment.appendChild(document.createElement("br"));
          continue;
        }
        if (lineOutput && lineOutput.length > 0) {
          for (let el of lineOutput) {
            try {
              MicronParser.upgradeImageLinks(el);
              fragment.appendChild(el);
            } catch (e) {
              console.warn("MicronParser: appendChild failed", e);
              const fallback = document.createElement("span");
              fallback.className = "mu-line-parse-fallback";
              fallback.style.whiteSpace = "pre-wrap";
              fallback.textContent = line;
              fragment.appendChild(fallback);
              fragment.appendChild(document.createElement("br"));
              break;
            }
          }
        } else if (lineOutput && lineOutput.length === 0) {
          // skip
        } else {
          fragment.appendChild(document.createElement("br"));
        }
      }

      return fragment;
    } catch (e) {
      console.warn("MicronParser: convertMicronToFragment failed", e);
      const fragment = document.createDocumentFragment();
      const pre = document.createElement("pre");
      pre.className = "mu-parse-fallback";
      pre.style.whiteSpace = "pre-wrap";
      pre.textContent = markup;
      fragment.appendChild(pre);
      return fragment;
    }
  }

  parseLine(line, state) {
    if (line.length > 0 && !state.literal) {
      const partialMatch = line.trim().match(MicronParser.PARTIAL_LINE_REGEX);
      if (partialMatch) {
        const dest = partialMatch[1];
        const path = partialMatch[2];
        const refresh = partialMatch[3] ? parseInt(partialMatch[3], 10) : null;
        const fields = partialMatch[4] || null;
        const id = "partial-" + state.partialIndex++;
        const div = document.createElement("div");
        div.className = "mu-partial";
        div.setAttribute("data-partial-id", id);
        div.setAttribute("data-dest", dest);
        div.setAttribute("data-path", path);
        if (refresh != null && refresh > 0) {
          div.setAttribute("data-refresh", String(refresh));
        }
        if (fields) {
          div.setAttribute("data-fields", fields);
        }
        div.textContent = "Loading...";
        return [div];
      }
    }
    return super.parseLine(line, state);
  }

  wrapWord(word) {
    if (word.length === 0) return "";
    if (!MicronParser.lineNeedsPerCharCells(word)) {
      return "<span class='Mu-mnt-group'>" + escapeHtmlForFallback(word) + "</span>";
    }
    let out = "";
    let charArr;
    try {
      charArr = [...new Intl.Segmenter().segment(word)].map((x) => x.segment);
    } catch {
      try {
        charArr = Array.from(word);
      } catch {
        charArr = word.split("");
      }
    }
    for (let char of charArr) {
      const cellClass = MicronParser.isWideMonospaceCell(char) ? "Mu-mnt-full" : "Mu-mnt";
      out += "<span class='" + cellClass + "'>" + escapeHtmlForFallback(char) + "</span>";
    }
    return "<span class='Mu-mws'>" + out + "</span>";
  }

  splitAtSpaces(line) {
    let out = "";
    const wordArr = line.split(/(?<= )/g);
    for (const word of wordArr) {
      out += this.wrapWord(word);
    }
    return out;
  }

  forceMonospace(line) {
    if (line == null || line === "") {
      return "";
    }
    if (!MicronParser.lineNeedsPerCharCells(line)) {
      return "<span class='Mu-mnt-group'>" + escapeHtmlForFallback(line) + "</span>";
    }
    let out = "";
    let charArr;
    try {
      charArr = [...new Intl.Segmenter().segment(line)].map((x) => x.segment);
    } catch {
      try {
        charArr = Array.from(line);
      } catch {
        charArr = line.split("");
      }
    }
    for (let char of charArr) {
      const cellClass = MicronParser.isWideMonospaceCell(char) ? "Mu-mnt-full" : "Mu-mnt";
      out += "<span class='" + cellClass + "'>" + escapeHtmlForFallback(char) + "</span>";
    }
    return out;
  }
}

export function parseMicronHeaderTagColors(
  source: string,
  darkTheme = true,
): { fg: string; bg: string } {
  const parser = new MicronParser(darkTheme, false);
  const colors = parser.parseHeaderTags(source) as { fg?: string | null; bg?: string | null };
  return { fg: colors.fg ?? "", bg: colors.bg ?? "" };
}

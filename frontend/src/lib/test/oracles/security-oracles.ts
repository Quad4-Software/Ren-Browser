// SPDX-License-Identifier: MIT

/** Shared oracles for exploratory / adversarial frontend tests. */

export const XSS_HTML_CORPUS: string[] = [
  `<script>window.__xss=1</script>`,
  `<img src=x onerror=alert(1)>`,
  `<svg onload=alert(1)>`,
  `<a href="javascript:alert(1)">x</a>`,
  `<a href="JaVaScRiPt:alert(1)">x</a>`,
  `<iframe src="javascript:alert(1)">`,
  `<object data="javascript:alert(1)">`,
  `<embed src="javascript:alert(1)">`,
  `<form action="https://evil.example"><input name=q></form>`,
  `<base href="https://evil.example/">`,
  `<meta http-equiv="refresh" content="0;url=https://evil.example">`,
  `<link rel="stylesheet" href="https://evil.example/x.css">`,
  `<div style="background:url(javascript:alert(1))">`,
  `<div style="background:url('https://evil.example/x')">`,
  `<style>@import url(https://evil.example/x.css);</style>`,
  `<img src="data:text/html,<script>alert(1)</script>">`,
  `<img src="https://evil.example/pwn.png">`,
  `<<script>alert(1)//<</script>`,
  `<a href="#" onclick="alert(1)">x</a>`,
  `<body onload=alert(1)>`,
  `<input onfocus=alert(1) autofocus>`,
  `<math><mtext><table><mglyph><style><!--</style><img title="--&gt;&lt;img src=x onerror=alert(1)&gt;">`,
];

export const CSS_BREAKOUT_COLORS: string[] = [
  `000;}</style><script>window.__xss=1</script><style>x{color:`,
  `fff" onload="alert(1)`,
  `abc';}</style><img src=x onerror=alert(1)><style>`,
  `red`,
  `url(https://evil.example)`,
  `expression(alert(1))`,
  `000";x`,
  `</style>`,
  `#fff;}</style><script>`,
  `000\n}</style><script>`,
];

export const NON_HEX_HEADER_COLORS: string[] = [
  "zzz",
  "ggg",
  "12g",
  '000";x',
  "fff on",
  "!!!!!!",
  "abcde!",
];

export const BLOCKED_NAV_CORPUS: string[] = [
  "https://example.com",
  "http://example.com/path",
  "HTTPS://EXAMPLE.COM",
  "//cdn.example.com",
  "ftp://files.example.com",
  "file:///etc/passwd",
  "javascript:alert(1)",
  "JaVaScRiPt:alert(1)",
  "data:text/html,<script>alert(1)</script>",
  "mailto:user@example.com",
  "tel:+15551212",
  "blob:https://example.com/uuid",
  "vbscript:msgbox(1)",
  "ws://example.com",
  "wss://example.com/socket",
  "http:example.com",
];

export const ALLOWED_NAV_CORPUS: string[] = [
  "about:",
  "license:",
  "editor:",
  "config:",
  "settings:",
  "docs:",
  "docs:?lang=en",
  "document:/book.epub",
  "abb3ebcd03cb2388a838e70c001291f9:/page/index.mu",
  "rns://abb3ebcd03cb2388a838e70c001291f9/page/index.mu",
  "ABB3EBCD03CB2388A838E70C001291F9",
  "/page/guide.mu",
];

export const DOWNLOAD_NAME_ATTACKS: Array<{ url: string; contentType: string }> = [
  { url: "abb3ebcd03cb2388a838e70c001291f9:/page/x.mu?a=../../etc/passwd", contentType: "micron" },
  {
    url: "abb3ebcd03cb2388a838e70c001291f9:/page/x.mu?a=..\\..\\windows\\win.ini",
    contentType: "micron",
  },
  { url: "abb3ebcd03cb2388a838e70c001291f9:/page/x.mu?a=/etc/passwd", contentType: "micron" },
  {
    url: "abb3ebcd03cb2388a838e70c001291f9:/page/x.mu?a=C:\\Windows\\System32\\config",
    contentType: "micron",
  },
  { url: "abb3ebcd03cb2388a838e70c001291f9:/page/x.mu?a=evil:payload.txt", contentType: "micron" },
  { url: "abb3ebcd03cb2388a838e70c001291f9:/page/x.mu?a=name%00.exe", contentType: "micron" },
  {
    url: "abb3ebcd03cb2388a838e70c001291f9:/page/x.mu?a=....//....//etc/passwd",
    contentType: "micron",
  },
];

const EVENT_HANDLER_ATTR_RE = /\son[a-z]+\s*=/i;
const SCRIPT_TAG_RE = /<\s*script[\s>/]/i;
const JS_URI_RE = /(?:^|[\s"'=])javascript\s*:/i;
const VB_URI_RE = /(?:^|[\s"'=])vbscript\s*:/i;
const DATA_HTML_URI_RE = /data\s*:\s*text\/html/i;

export function assertNoExecutableMarkup(html: string, label: string): void {
  const lower = html.toLowerCase();
  if (SCRIPT_TAG_RE.test(html)) {
    throw new Error(`oracle(${label}): script tag survived\n${html.slice(0, 400)}`);
  }
  if (EVENT_HANDLER_ATTR_RE.test(html)) {
    throw new Error(`oracle(${label}): event handler attribute survived\n${html.slice(0, 400)}`);
  }
  if (JS_URI_RE.test(html) || VB_URI_RE.test(html)) {
    throw new Error(`oracle(${label}): script URI survived\n${html.slice(0, 400)}`);
  }
  if (DATA_HTML_URI_RE.test(lower)) {
    throw new Error(`oracle(${label}): data:text/html URI survived\n${html.slice(0, 400)}`);
  }
  for (const tag of ["iframe", "object", "embed", "base", "meta", "form", "link"]) {
    if (new RegExp(`<\\s*${tag}\\b`, "i").test(html)) {
      throw new Error(`oracle(${label}): forbidden <${tag}> survived\n${html.slice(0, 400)}`);
    }
  }
}

export function assertCssColorTokenSafe(value: string, label: string): void {
  const trimmed = value.trim();
  if (!trimmed) {
    return;
  }
  if (!/^#[0-9a-fA-F]{3}$|^#[0-9a-fA-F]{6}$|^#[0-9a-fA-F]{8}$/.test(trimmed)) {
    throw new Error(`oracle(${label}): unsafe CSS color token ${JSON.stringify(value)}`);
  }
}

export function assertNoStyleBreakout(doc: string, label: string): void {
  const lower = doc.toLowerCase();
  const styleClose = lower.indexOf("</style>");
  if (styleClose < 0) {
    return;
  }
  const after = lower.slice(styleClose + "</style>".length);
  const nextStyle = after.indexOf("<style");
  const bodyOpen = after.indexOf("<body");
  const end = nextStyle >= 0 ? nextStyle : after.length;
  const between = after.slice(0, end);
  if (/<\s*(script|img|svg|iframe|object|embed|link|meta|base)\b/i.test(between)) {
    throw new Error(`oracle(${label}): markup breakout after </style>\n${doc.slice(0, 500)}`);
  }
  if (bodyOpen >= 0 && bodyOpen < end && /on[a-z]+\s*=/i.test(between.slice(0, bodyOpen + 80))) {
    throw new Error(`oracle(${label}): attribute breakout near body\n${doc.slice(0, 500)}`);
  }
}

export function assertSafeDownloadFilename(name: string, label: string): void {
  if (!name || name === "." || name === "..") {
    throw new Error(`oracle(${label}): empty or dot filename ${JSON.stringify(name)}`);
  }
  if (name.includes("/") || name.includes("\\") || name.includes("\0")) {
    throw new Error(`oracle(${label}): path separator or NUL in ${JSON.stringify(name)}`);
  }
  if (name.includes("..")) {
    throw new Error(`oracle(${label}): parent traversal in ${JSON.stringify(name)}`);
  }
  if (/^[a-zA-Z]:/.test(name) || name.includes(":")) {
    throw new Error(`oracle(${label}): drive or scheme-like filename ${JSON.stringify(name)}`);
  }
}

export function mulberry32(seed: number): () => number {
  let t = seed >>> 0;
  return () => {
    t += 0x6d2b79f5;
    let r = Math.imul(t ^ (t >>> 15), 1 | t);
    r ^= r + Math.imul(r ^ (r >>> 7), 61 | r);
    return ((r ^ (r >>> 14)) >>> 0) / 4294967296;
  };
}

export function randomHexish(rand: () => number, len: 3 | 6): string {
  const alphabet = "0123456789abcdefABCDEFzzzz!!!!\"';<>/";
  let out = "";
  for (let i = 0; i < len; i++) {
    out += alphabet[Math.floor(rand() * alphabet.length)]!;
  }
  return out;
}

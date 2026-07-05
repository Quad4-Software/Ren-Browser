// SPDX-License-Identifier: MIT
import { ResolveMicronLink } from "../../../bindings/renbrowser/internal/app/browserservice.js";
import { normalizeReticulumURL } from "./url";

export type FieldInput = {
  type: string;
  name: string;
  value: string;
  checked: boolean;
};

export function collectFormInputs(root: ParentNode): FieldInput[] {
  const inputs = root.querySelectorAll<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>(
    "input, textarea, select",
  );
  const out: FieldInput[] = [];
  for (const el of inputs) {
    if (el instanceof HTMLInputElement) {
      out.push({
        type: el.type || "text",
        name: el.name,
        value: el.value,
        checked: el.checked,
      });
    } else if (el instanceof HTMLTextAreaElement) {
      out.push({
        type: "textarea",
        name: el.name,
        value: el.value,
        checked: false,
      });
    } else if (el instanceof HTMLSelectElement) {
      out.push({
        type: "select",
        name: el.name,
        value: el.value,
        checked: false,
      });
    }
  }
  return out;
}

export async function resolveMicronNavigation(
  root: ParentNode,
  currentURL: string,
  destination: string,
  fieldsSpec: string | null,
): Promise<string> {
  const inputs = collectFormInputs(root);
  const spec = fieldsSpec ?? "";
  const next = await ResolveMicronLink(currentURL, destination, spec, inputs);
  return normalizeReticulumURL(next);
}

export function resolveLinkURL(currentURL: string, href: string): string {
  const trimmed = href.trim();
  if (!trimmed) {
    return "";
  }
  const special = normalizeReticulumURL(trimmed);
  if (special === "about:" || special === "license:" || special === "editor:" || special === "config:" || special.startsWith("docs")) {
    return special.startsWith("docs") ? normalizeReticulumURL(trimmed) : special;
  }
  if (trimmed.includes(":/")) {
    return normalizeReticulumURL(trimmed);
  }

  const sep = currentURL.indexOf(":/");
  if (sep < 0) {
    return "";
  }
  const node = currentURL.slice(0, sep);
  const currentRest = currentURL.slice(sep + 2);
  const currentPath = currentRest.split(/[?`]/)[0] ?? "";
  const pathBase = currentPath.startsWith("/") ? currentPath : `/${currentPath}`;

  if (trimmed.startsWith("?") || trimmed.startsWith("`")) {
    return normalizeReticulumURL(`${node}:${pathBase}${trimmed}`);
  }

  let path = trimmed;
  if (!path.startsWith("/")) {
    path = "/" + path;
  }
  if (!path.startsWith("/page/") && !path.startsWith("/file/")) {
    path = "/page/" + path.replace(/^\//, "");
  }
  return `${node}:${path}`;
}

export function resolveNomadDataURL(currentURL: string, dataUrl: string): string {
  return normalizeReticulumURL(
    dataUrl.includes(":/") ? dataUrl : resolveLinkURL(currentURL, dataUrl),
  );
}

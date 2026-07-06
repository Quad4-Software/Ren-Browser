// SPDX-License-Identifier: MIT
import { resolveLinkURL, resolveMicronNavigation, resolveNomadDataURL } from "./micron-links";
import { isBlockedExternalHref } from "./navigation-guard";

export async function handlePageLinkClick(
  event: MouseEvent,
  root: ParentNode,
  currentURL: string,
  onNavigate: (url: string) => void | Promise<void>,
): Promise<void> {
  const target = event.target;
  if (!(target instanceof HTMLElement)) {
    return;
  }

  const nodeLink = target.closest("[data-action='openNode']");
  if (nodeLink) {
    event.preventDefault();
    const destination = nodeLink.getAttribute("data-destination");
    if (!destination) {
      throw new Error("Link is missing a destination");
    }
    const fieldsSpec = nodeLink.getAttribute("data-fields");
    const next = await resolveMicronNavigation(root, currentURL, destination, fieldsSpec);
    if (!next) {
      throw new Error(`Could not resolve link destination: ${destination}`);
    }
    await onNavigate(next);
    return;
  }

  const nomadAnchor = target.closest("a[data-nomad-url]");
  if (nomadAnchor) {
    event.preventDefault();
    const dataUrl = nomadAnchor.getAttribute("data-nomad-url");
    if (!dataUrl) {
      throw new Error("Link is missing a destination");
    }
    await onNavigate(resolveNomadDataURL(currentURL, dataUrl));
    return;
  }

  const anchor = target.closest("a");
  if (!anchor) {
    return;
  }

  event.preventDefault();

  const href = anchor.getAttribute("href");
  if (!href || isBlockedExternalHref(href)) {
    return;
  }

  if (href.startsWith("#")) {
    const id = href.slice(1);
    if (id) {
      root.querySelector(`#${CSS.escape(id)}`)?.scrollIntoView({ block: "start" });
    }
    return;
  }

  const next = resolveLinkURL(currentURL, href);
  if (!next) {
    throw new Error(`Could not resolve link: ${href}`);
  }
  await onNavigate(next);
}

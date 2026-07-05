// SPDX-License-Identifier: MIT
export function clearFindHighlights(root: HTMLElement): void {
  root.querySelectorAll("mark.ren-find-hit").forEach((mark) => {
    const parent = mark.parentNode;
    if (!parent) {
      return;
    }
    parent.replaceChild(document.createTextNode(mark.textContent ?? ""), mark);
    parent.normalize();
  });
}

export function highlightFindMatches(root: HTMLElement, query: string): number {
  clearFindHighlights(root);
  const needle = query.trim();
  if (!needle) {
    return 0;
  }
  const lower = needle.toLowerCase();
  let count = 0;
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
  const textNodes: Text[] = [];
  while (walker.nextNode()) {
    const node = walker.currentNode as Text;
    if (!node.nodeValue || !node.parentElement) {
      continue;
    }
    if (node.parentElement.closest("mark.ren-find-hit")) {
      continue;
    }
    textNodes.push(node);
  }

  for (const node of textNodes) {
    const value = node.nodeValue ?? "";
    const valueLower = value.toLowerCase();
    let start = 0;
    let index = valueLower.indexOf(lower, start);
    if (index < 0) {
      continue;
    }
    const frag = document.createDocumentFragment();
    while (index >= 0) {
      if (index > start) {
        frag.appendChild(document.createTextNode(value.slice(start, index)));
      }
      const mark = document.createElement("mark");
      mark.className = "ren-find-hit";
      mark.textContent = value.slice(index, index + needle.length);
      frag.appendChild(mark);
      count++;
      start = index + needle.length;
      index = valueLower.indexOf(lower, start);
    }
    if (start < value.length) {
      frag.appendChild(document.createTextNode(value.slice(start)));
    }
    node.parentNode?.replaceChild(frag, node);
  }
  return count;
}

export function scrollToFindMatch(root: HTMLElement, index: number): void {
  const marks = root.querySelectorAll<HTMLElement>("mark.ren-find-hit");
  if (!marks.length) {
    return;
  }
  const clamped = ((index % marks.length) + marks.length) % marks.length;
  marks.forEach((mark, i) => {
    mark.classList.toggle("ren-find-active", i === clamped);
  });
  marks[clamped]?.scrollIntoView({ block: "center", behavior: "smooth" });
}
